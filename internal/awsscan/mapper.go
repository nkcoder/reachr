package awsscan

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/nkcoder/reachr/internal/topology"
)

// ToSnapshot maps the raw AWS backbone into the normalized topology.json schema.
// It is pure (no AWS calls) and performs no reachability reasoning — only the
// deterministic re-encoding decided in internal/topology/doc.go.
func (r *RawBackbone) ToSnapshot(account, filter string) *topology.Snapshot {
	snap := &topology.Snapshot{
		SchemaVersion: topology.SchemaVersion,
		ScannedAt:     time.Now().UTC(),
		Region:        r.Region,
		Account:       account,
		Filter:        filter,
	}
	for _, v := range r.VPCs {
		snap.VPCs = append(snap.VPCs, mapVPC(v))
	}
	for _, s := range r.Subnets {
		snap.Subnets = append(snap.Subnets, mapSubnet(s))
	}
	for _, rt := range r.RouteTables {
		snap.RouteTables = append(snap.RouteTables, mapRouteTable(rt))
	}
	for _, ig := range r.InternetGateways {
		snap.InternetGateways = append(snap.InternetGateways, mapIGW(ig))
	}
	for _, n := range r.NATGateways {
		snap.NATGateways = append(snap.NATGateways, mapNAT(n))
	}
	for _, e := range r.NetworkInterfaces {
		snap.ENIs = append(snap.ENIs, mapENI(e))
	}
	for _, sg := range r.SecurityGroups {
		snap.SecurityGroups = append(snap.SecurityGroups, mapSG(sg))
	}
	for _, ve := range r.VPCEndpoints {
		snap.VPCEndpoints = append(snap.VPCEndpoints, mapEndpoint(ve))
	}
	return snap
}

func mapVPC(v types.Vpc) topology.VPC {
	out := topology.VPC{
		ID:        aws.ToString(v.VpcId),
		IsDefault: aws.ToBool(v.IsDefault),
		State:     string(v.State),
		Tags:      tagMap(v.Tags),
	}
	for _, a := range v.CidrBlockAssociationSet {
		out.CIDRBlocks = append(out.CIDRBlocks, aws.ToString(a.CidrBlock))
	}
	for _, a := range v.Ipv6CidrBlockAssociationSet {
		out.IPv6CIDRs = append(out.IPv6CIDRs, aws.ToString(a.Ipv6CidrBlock))
	}
	return out
}

func mapSubnet(s types.Subnet) topology.Subnet {
	out := topology.Subnet{
		ID:                  aws.ToString(s.SubnetId),
		VPCID:               aws.ToString(s.VpcId),
		CIDRBlock:           aws.ToString(s.CidrBlock),
		AvailabilityZone:    aws.ToString(s.AvailabilityZone),
		AvailabilityZoneID:  aws.ToString(s.AvailabilityZoneId),
		MapPublicIPOnLaunch: aws.ToBool(s.MapPublicIpOnLaunch),
		ARN:                 aws.ToString(s.SubnetArn),
		Tags:                tagMap(s.Tags),
	}
	if len(s.Ipv6CidrBlockAssociationSet) > 0 {
		out.IPv6CIDR = aws.ToString(s.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}
	return out
}

func mapRouteTable(rt types.RouteTable) topology.RouteTable {
	out := topology.RouteTable{
		ID:    aws.ToString(rt.RouteTableId),
		VPCID: aws.ToString(rt.VpcId),
		Tags:  tagMap(rt.Tags),
	}
	for _, a := range rt.Associations {
		if aws.ToBool(a.Main) {
			out.Main = true
		}
		if a.SubnetId != nil {
			out.SubnetIDs = append(out.SubnetIDs, aws.ToString(a.SubnetId))
		}
	}
	for _, rte := range rt.Routes {
		out.Routes = append(out.Routes, mapRoute(rte))
	}
	return out
}

func mapRoute(r types.Route) topology.Route {
	return topology.Route{
		Destination: routeDestination(r),
		Target:      routeTarget(r),
		State:       string(r.State),
		Origin:      string(r.Origin),
	}
}

func routeDestination(r types.Route) topology.RouteDestination {
	switch {
	case r.DestinationCidrBlock != nil:
		return topology.RouteDestination{Kind: topology.DestCIDR, Value: aws.ToString(r.DestinationCidrBlock)}
	case r.DestinationIpv6CidrBlock != nil:
		return topology.RouteDestination{Kind: topology.DestIPv6CIDR, Value: aws.ToString(r.DestinationIpv6CidrBlock)}
	case r.DestinationPrefixListId != nil:
		return topology.RouteDestination{Kind: topology.DestPrefixList, Value: aws.ToString(r.DestinationPrefixListId)}
	}
	return topology.RouteDestination{}
}

// routeTarget resolves the single non-nil target field. GatewayId is overloaded
// ("local" / igw-* / vpce-*), so it is sub-typed by id prefix.
func routeTarget(r types.Route) topology.RouteTarget {
	switch {
	case r.GatewayId != nil:
		id := aws.ToString(r.GatewayId)
		switch {
		case id == "local":
			return topology.RouteTarget{Kind: topology.TargetLocal, ID: id}
		case strings.HasPrefix(id, "igw-"):
			return topology.RouteTarget{Kind: topology.TargetIGW, ID: id}
		case strings.HasPrefix(id, "vpce-"):
			return topology.RouteTarget{Kind: topology.TargetVPCEndpointGW, ID: id}
		default:
			return topology.RouteTarget{Kind: topology.TargetUnknown, ID: id}
		}
	case r.NatGatewayId != nil:
		return topology.RouteTarget{Kind: topology.TargetNAT, ID: aws.ToString(r.NatGatewayId)}
	case r.NetworkInterfaceId != nil:
		return topology.RouteTarget{Kind: topology.TargetENI, ID: aws.ToString(r.NetworkInterfaceId)}
	case r.TransitGatewayId != nil:
		return topology.RouteTarget{Kind: topology.TargetTGW, ID: aws.ToString(r.TransitGatewayId)}
	case r.VpcPeeringConnectionId != nil:
		return topology.RouteTarget{Kind: topology.TargetPeering, ID: aws.ToString(r.VpcPeeringConnectionId)}
	case r.EgressOnlyInternetGatewayId != nil:
		return topology.RouteTarget{Kind: topology.TargetEgressOnlyIGW, ID: aws.ToString(r.EgressOnlyInternetGatewayId)}
	case r.InstanceId != nil:
		return topology.RouteTarget{Kind: topology.TargetInstance, ID: aws.ToString(r.InstanceId)}
	}
	return topology.RouteTarget{Kind: topology.TargetUnknown}
}

func mapIGW(ig types.InternetGateway) topology.InternetGateway {
	out := topology.InternetGateway{
		ID:   aws.ToString(ig.InternetGatewayId),
		Tags: tagMap(ig.Tags),
	}
	if len(ig.Attachments) > 0 {
		out.AttachedVPC = aws.ToString(ig.Attachments[0].VpcId)
	}
	return out
}

func mapNAT(n types.NatGateway) topology.NATGateway {
	return topology.NATGateway{
		ID:       aws.ToString(n.NatGatewayId),
		VPCID:    aws.ToString(n.VpcId),
		SubnetID: aws.ToString(n.SubnetId),
		State:    string(n.State),
		Tags:     tagMap(n.Tags),
	}
}

func mapENI(e types.NetworkInterface) topology.ENI {
	out := topology.ENI{
		ID:               aws.ToString(e.NetworkInterfaceId),
		VPCID:            aws.ToString(e.VpcId),
		SubnetID:         aws.ToString(e.SubnetId),
		AvailabilityZone: aws.ToString(e.AvailabilityZone),
		Description:      aws.ToString(e.Description),
		InterfaceType:    string(e.InterfaceType),
		Status:           string(e.Status),
		Tags:             tagMap(e.TagSet),
	}
	for _, ip := range e.PrivateIpAddresses {
		out.PrivateIPs = append(out.PrivateIPs, aws.ToString(ip.PrivateIpAddress))
	}
	if e.Association != nil {
		out.PublicIP = aws.ToString(e.Association.PublicIp)
	}
	for _, g := range e.Groups {
		out.SecurityGroupIDs = append(out.SecurityGroupIDs, aws.ToString(g.GroupId))
	}
	if e.Attachment != nil {
		out.Attachment = &topology.ENIAttachment{
			InstanceID:       aws.ToString(e.Attachment.InstanceId),
			InstanceOwnerID:  aws.ToString(e.Attachment.InstanceOwnerId),
			RequesterID:      aws.ToString(e.RequesterId),
			RequesterManaged: aws.ToBool(e.RequesterManaged),
			DeviceIndex:      aws.ToInt32(e.Attachment.DeviceIndex),
			Status:           string(e.Attachment.Status),
		}
	}
	return out
}

func mapSG(sg types.SecurityGroup) topology.SecurityGroup {
	out := topology.SecurityGroup{
		ID:          aws.ToString(sg.GroupId),
		VPCID:       aws.ToString(sg.VpcId),
		Name:        aws.ToString(sg.GroupName),
		Description: aws.ToString(sg.Description),
		ARN:         aws.ToString(sg.SecurityGroupArn),
		Tags:        tagMap(sg.Tags),
	}
	for _, p := range sg.IpPermissions {
		out.Rules = append(out.Rules, flattenPermission(topology.Ingress, p)...)
	}
	for _, p := range sg.IpPermissionsEgress {
		out.Rules = append(out.Rules, flattenPermission(topology.Egress, p)...)
	}
	return out
}

// flattenPermission explodes one AWS permission into one atomic rule per peer.
func flattenPermission(dir topology.RuleDirection, p types.IpPermission) []topology.SGRule {
	proto := aws.ToString(p.IpProtocol)
	if proto == "-1" {
		proto = "all"
	}
	base := topology.SGRule{
		Direction: dir,
		Protocol:  proto,
		FromPort:  p.FromPort,
		ToPort:    p.ToPort,
	}
	var rules []topology.SGRule
	rule := func(peer topology.SGPeer, desc string) {
		r := base
		r.Peer = peer
		r.Description = desc
		rules = append(rules, r)
	}
	for _, c := range p.IpRanges {
		rule(topology.SGPeer{Kind: topology.PeerCIDR, Value: aws.ToString(c.CidrIp)}, aws.ToString(c.Description))
	}
	for _, c := range p.Ipv6Ranges {
		rule(topology.SGPeer{Kind: topology.PeerIPv6, Value: aws.ToString(c.CidrIpv6)}, aws.ToString(c.Description))
	}
	for _, pl := range p.PrefixListIds {
		rule(topology.SGPeer{Kind: topology.PeerPrefixList, Value: aws.ToString(pl.PrefixListId)}, aws.ToString(pl.Description))
	}
	for _, g := range p.UserIdGroupPairs {
		rule(topology.SGPeer{Kind: topology.PeerSG, Value: aws.ToString(g.GroupId), Account: aws.ToString(g.UserId)}, aws.ToString(g.Description))
	}
	return rules
}

func mapEndpoint(ve types.VpcEndpoint) topology.VPCEndpoint {
	out := topology.VPCEndpoint{
		ID:                aws.ToString(ve.VpcEndpointId),
		VPCID:             aws.ToString(ve.VpcId),
		ServiceName:       aws.ToString(ve.ServiceName),
		Kind:              endpointKind(ve.VpcEndpointType),
		State:             string(ve.State),
		RouteTableIDs:     ve.RouteTableIds,
		SubnetIDs:         ve.SubnetIds,
		ENIIDs:            ve.NetworkInterfaceIds,
		PrivateDNSEnabled: aws.ToBool(ve.PrivateDnsEnabled),
		Tags:              tagMap(ve.Tags),
	}
	for _, g := range ve.Groups {
		out.SecurityGroupIDs = append(out.SecurityGroupIDs, aws.ToString(g.GroupId))
	}
	return out
}

func endpointKind(t types.VpcEndpointType) topology.EndpointKind {
	if t == types.VpcEndpointTypeInterface {
		return topology.EndpointInterface
	}
	return topology.EndpointGateway
}

// tagMap normalizes AWS []Tag (or nil) into a map for O(1) scope-selector lookups.
func tagMap(tags []types.Tag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}
	return m
}
