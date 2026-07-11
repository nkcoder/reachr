package awsscan

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// RawBackbone is the unshaped exploratory dump of the network backbone, straight
// from EC2 Describe* calls. It exists to inspect real field shapes before locking
// the topology.json schema (issue #6); it is NOT the final snapshot format.
type RawBackbone struct {
	Region            string                   `json:"region"`
	VPCs              []types.Vpc              `json:"vpcs"`
	Subnets           []types.Subnet           `json:"subnets"`
	RouteTables       []types.RouteTable       `json:"routeTables"`
	InternetGateways  []types.InternetGateway  `json:"internetGateways"`
	NATGateways       []types.NatGateway       `json:"natGateways"`
	NetworkInterfaces []types.NetworkInterface `json:"networkInterfaces"`
	SecurityGroups    []types.SecurityGroup    `json:"securityGroups"`
}

// paginator is the shape shared by every EC2 Describe*Paginator.
type paginator[T any] interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*ec2.Options)) (T, error)
}

// drain walks a paginator to completion, flattening each page via extract.
func drain[Page, Item any](ctx context.Context, p paginator[Page], extract func(Page) []Item) ([]Item, error) {
	var all []Item
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		all = append(all, extract(page)...)
	}
	return all, nil
}

// DumpBackbone collects the raw network backbone with pagination.
func (c *Client) DumpBackbone(ctx context.Context) (*RawBackbone, error) {
	out := &RawBackbone{Region: c.Region}
	var err error

	if out.VPCs, err = drain(ctx,
		ec2.NewDescribeVpcsPaginator(c.EC2, &ec2.DescribeVpcsInput{}),
		func(p *ec2.DescribeVpcsOutput) []types.Vpc { return p.Vpcs }); err != nil {
		return nil, err
	}
	if out.Subnets, err = drain(ctx,
		ec2.NewDescribeSubnetsPaginator(c.EC2, &ec2.DescribeSubnetsInput{}),
		func(p *ec2.DescribeSubnetsOutput) []types.Subnet { return p.Subnets }); err != nil {
		return nil, err
	}
	if out.RouteTables, err = drain(ctx,
		ec2.NewDescribeRouteTablesPaginator(c.EC2, &ec2.DescribeRouteTablesInput{}),
		func(p *ec2.DescribeRouteTablesOutput) []types.RouteTable { return p.RouteTables }); err != nil {
		return nil, err
	}
	if out.InternetGateways, err = drain(ctx,
		ec2.NewDescribeInternetGatewaysPaginator(c.EC2, &ec2.DescribeInternetGatewaysInput{}),
		func(p *ec2.DescribeInternetGatewaysOutput) []types.InternetGateway { return p.InternetGateways }); err != nil {
		return nil, err
	}
	if out.NATGateways, err = drain(ctx,
		ec2.NewDescribeNatGatewaysPaginator(c.EC2, &ec2.DescribeNatGatewaysInput{}),
		func(p *ec2.DescribeNatGatewaysOutput) []types.NatGateway { return p.NatGateways }); err != nil {
		return nil, err
	}
	if out.NetworkInterfaces, err = drain(ctx,
		ec2.NewDescribeNetworkInterfacesPaginator(c.EC2, &ec2.DescribeNetworkInterfacesInput{}),
		func(p *ec2.DescribeNetworkInterfacesOutput) []types.NetworkInterface { return p.NetworkInterfaces }); err != nil {
		return nil, err
	}
	if out.SecurityGroups, err = drain(ctx,
		ec2.NewDescribeSecurityGroupsPaginator(c.EC2, &ec2.DescribeSecurityGroupsInput{}),
		func(p *ec2.DescribeSecurityGroupsOutput) []types.SecurityGroup { return p.SecurityGroups }); err != nil {
		return nil, err
	}
	return out, nil
}
