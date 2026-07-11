// Package awsscan collects the AWS network backbone (read-only) for a scan.
package awsscan

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// Client wraps the AWS service clients used for scanning.
type Client struct {
	EC2    *ec2.Client
	Region string
}

// NewClient builds a Client from the AWS default credential chain.
// region is required; profile is optional (empty = default chain).
func NewClient(ctx context.Context, region, profile string) (*Client, error) {
	opts := []func(*config.LoadOptions) error{config.WithRegion(region)}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{EC2: ec2.NewFromConfig(cfg), Region: cfg.Region}, nil
}
