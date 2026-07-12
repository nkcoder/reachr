package topology

import "time"

// SchemaVersion is the current topology.json schema version.
const SchemaVersion = 1

// Snapshot is the top-level immutable scan result — a pure function input for
// render/explain/diff. Collections are arrays; consumers index by id at load.
type Snapshot struct {
	SchemaVersion    int               `json:"schemaVersion"`
	ScannedAt        time.Time         `json:"scannedAt"`
	Region           string            `json:"region"`
	Account          string            `json:"account"`
	Filter           string            `json:"filter,omitempty"`
	ToolVersion      string            `json:"toolVersion,omitempty"`
	VPCs             []VPC             `json:"vpcs"`
	Subnets          []Subnet          `json:"subnets"`
	RouteTables      []RouteTable      `json:"routeTables"`
	InternetGateways []InternetGateway `json:"internetGateways"`
	NATGateways      []NATGateway      `json:"natGateways"`
	ENIs             []ENI             `json:"enis"`
	SecurityGroups   []SecurityGroup   `json:"securityGroups"`
	VPCEndpoints     []VPCEndpoint     `json:"vpcEndpoints"`
}

// VPC is a virtual private cloud. CIDRBlocks flattens CidrBlockAssociationSet.
type VPC struct {
	ID         string            `json:"id"`
	CIDRBlocks []string          `json:"cidrBlocks"`
	IPv6CIDRs  []string          `json:"ipv6Cidrs,omitempty"`
	IsDefault  bool              `json:"isDefault"`
	State      string            `json:"state"`
	Tags       map[string]string `json:"tags,omitempty"`
}

// Subnet is an AZ-scoped CIDR within a VPC. MapPublicIPOnLaunch is a public/private hint.
type Subnet struct {
	ID                  string            `json:"id"`
	VPCID               string            `json:"vpcId"`
	CIDRBlock           string            `json:"cidrBlock"`
	IPv6CIDR            string            `json:"ipv6Cidr,omitempty"`
	AvailabilityZone    string            `json:"availabilityZone"`
	AvailabilityZoneID  string            `json:"availabilityZoneId"`
	MapPublicIPOnLaunch bool              `json:"mapPublicIpOnLaunch"`
	ARN                 string            `json:"arn,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
}

// RouteTable holds routes and its subnet associations. Main is true when it is the
// VPC's main table (applies to subnets without an explicit association in SubnetIDs).
type RouteTable struct {
	ID        string            `json:"id"`
	VPCID     string            `json:"vpcId"`
	Main      bool              `json:"main"`
	SubnetIDs []string          `json:"subnetIds"`
	Routes    []Route           `json:"routes"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// Route is a normalized route entry: a destination, a resolved target, and its state.
type Route struct {
	Destination RouteDestination `json:"destination"`
	Target      RouteTarget      `json:"target"`
	State       string           `json:"state"` // active | blackhole
	Origin      string           `json:"origin,omitempty"`
}

// DestinationKind discriminates a route's destination.
type DestinationKind string

// Route destination kinds.
const (
	DestCIDR       DestinationKind = "cidr"
	DestIPv6CIDR   DestinationKind = "ipv6-cidr"
	DestPrefixList DestinationKind = "prefix-list"
)

// RouteDestination is the matched-against side of a route (longest-prefix by value).
type RouteDestination struct {
	Kind  DestinationKind `json:"kind"`
	Value string          `json:"value"`
}

// TargetKind discriminates a route's next hop. TargetUnknown preserves the id of an
// AWS target field we do not yet model, so the engine can flag it honestly.
type TargetKind string

// Route target kinds (next-hop discriminators).
const (
	TargetLocal         TargetKind = "local"
	TargetIGW           TargetKind = "igw"
	TargetNAT           TargetKind = "nat"
	TargetENI           TargetKind = "eni"
	TargetVPCEndpointGW TargetKind = "vpc-endpoint-gw"
	TargetPeering       TargetKind = "peering"
	TargetTGW           TargetKind = "tgw"
	TargetEgressOnlyIGW TargetKind = "egress-only-igw"
	TargetInstance      TargetKind = "instance"
	TargetUnknown       TargetKind = "unknown"
)

// RouteTarget is the resolved next hop.
type RouteTarget struct {
	Kind TargetKind `json:"kind"`
	ID   string     `json:"id,omitempty"`
}

// InternetGateway is an IGW and the VPC it is attached to, if any.
type InternetGateway struct {
	ID          string            `json:"id"`
	AttachedVPC string            `json:"attachedVpc,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// NATGateway is a NAT gateway in a subnet.
type NATGateway struct {
	ID       string            `json:"id"`
	VPCID    string            `json:"vpcId"`
	SubnetID string            `json:"subnetId"`
	State    string            `json:"state"`
	Tags     map[string]string `json:"tags,omitempty"`
}

// ENI is the graph backbone: everything network-attached hangs off one. Ownership
// signals (Attachment, Description, InterfaceType) are stored raw; classification to a
// logical resource happens in the render/aggregation layer.
type ENI struct {
	ID               string            `json:"id"`
	VPCID            string            `json:"vpcId"`
	SubnetID         string            `json:"subnetId"`
	AvailabilityZone string            `json:"availabilityZone"`
	Description      string            `json:"description,omitempty"`
	InterfaceType    string            `json:"interfaceType"`
	Status           string            `json:"status"`
	PrivateIPs       []string          `json:"privateIps"`
	PublicIP         string            `json:"publicIp,omitempty"`
	SecurityGroupIDs []string          `json:"securityGroupIds"`
	Attachment       *ENIAttachment    `json:"attachment,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`
}

// ENIAttachment holds the raw ownership signals. InstanceOwnerID distinguishes a
// customer instance (account id) from service-managed ENIs ("amazon-elb", "amazon-rds",
// "amazon-aws" for NAT, etc.); RequesterID/Description further identify the owner.
type ENIAttachment struct {
	InstanceID       string `json:"instanceId,omitempty"`
	InstanceOwnerID  string `json:"instanceOwnerId,omitempty"`
	RequesterID      string `json:"requesterId,omitempty"`
	RequesterManaged bool   `json:"requesterManaged"`
	DeviceIndex      int32  `json:"deviceIndex"`
	Status           string `json:"status"`
}

// SecurityGroup carries its rules flattened to atomic form.
type SecurityGroup struct {
	ID          string            `json:"id"`
	VPCID       string            `json:"vpcId"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	ARN         string            `json:"arn,omitempty"`
	Rules       []SGRule          `json:"rules"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// RuleDirection is a security-group rule direction.
type RuleDirection string

// Security-group rule directions.
const (
	Ingress RuleDirection = "ingress"
	Egress  RuleDirection = "egress"
)

// SGRule is one atomic rule: a direction, protocol, port range, and a single peer.
// Protocol "all" is AWS "-1". Nil FromPort/ToPort means all ports.
type SGRule struct {
	Direction   RuleDirection `json:"direction"`
	Protocol    string        `json:"protocol"`
	FromPort    *int32        `json:"fromPort,omitempty"`
	ToPort      *int32        `json:"toPort,omitempty"`
	Peer        SGPeer        `json:"peer"`
	Description string        `json:"description,omitempty"`
}

// PeerKind discriminates a rule's peer.
type PeerKind string

// Security-group rule peer kinds.
const (
	PeerCIDR       PeerKind = "cidr"
	PeerIPv6       PeerKind = "ipv6"
	PeerSG         PeerKind = "sg"
	PeerPrefixList PeerKind = "prefix-list"
)

// SGPeer is the other side of a rule. Account is set for cross-account SG references.
type SGPeer struct {
	Kind    PeerKind `json:"kind"`
	Value   string   `json:"value"`
	Account string   `json:"account,omitempty"`
}

// EndpointKind discriminates a VPC endpoint. Gateway endpoints attach to route tables
// (no ENI); interface endpoints ride the ENI+SG model.
type EndpointKind string

// VPC endpoint kinds.
const (
	EndpointGateway   EndpointKind = "gateway"
	EndpointInterface EndpointKind = "interface"
)

// VPCEndpoint is a gateway or interface VPC endpoint (PrivateLink near-side).
type VPCEndpoint struct {
	ID                string            `json:"id"`
	VPCID             string            `json:"vpcId"`
	ServiceName       string            `json:"serviceName"`
	Kind              EndpointKind      `json:"kind"`
	State             string            `json:"state"`
	RouteTableIDs     []string          `json:"routeTableIds,omitempty"`
	SubnetIDs         []string          `json:"subnetIds,omitempty"`
	ENIIDs            []string          `json:"eniIds,omitempty"`
	SecurityGroupIDs  []string          `json:"securityGroupIds,omitempty"`
	PrivateDNSEnabled bool              `json:"privateDnsEnabled"`
	Tags              map[string]string `json:"tags,omitempty"`
}
