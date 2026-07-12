package topology

import "testing"

func TestSanitizeRedactsAccount(t *testing.T) {
	const acct = "123456789012"
	s := &Snapshot{
		Account: acct,
		Subnets: []Subnet{{ARN: "arn:aws:ec2:ap-southeast-2:" + acct + ":subnet/subnet-1"}},
		SecurityGroups: []SecurityGroup{{
			ARN: "arn:aws:ec2:ap-southeast-2:" + acct + ":security-group/sg-1",
			Rules: []SGRule{
				{Peer: SGPeer{Kind: PeerSG, Value: "sg-2", Account: acct}},
				{Peer: SGPeer{Kind: PeerCIDR, Value: "10.0.0.0/8"}},
			},
		}},
		ENIs: []ENI{
			{Attachment: &ENIAttachment{InstanceOwnerID: acct}},        // customer instance
			{Attachment: &ENIAttachment{InstanceOwnerID: "amazon-rds"}}, // service-managed
		},
	}

	Sanitize(s)

	if s.Account != RedactedAccount {
		t.Errorf("Account = %q, want %q", s.Account, RedactedAccount)
	}
	if got := s.Subnets[0].ARN; got != "arn:aws:ec2:ap-southeast-2:"+RedactedAccount+":subnet/subnet-1" {
		t.Errorf("subnet ARN not redacted: %q", got)
	}
	if got := s.SecurityGroups[0].Rules[0].Peer.Account; got != RedactedAccount {
		t.Errorf("SG-peer account = %q, want redacted", got)
	}
	if got := s.ENIs[0].Attachment.InstanceOwnerID; got != RedactedAccount {
		t.Errorf("instance owner = %q, want redacted", got)
	}
	if got := s.ENIs[1].Attachment.InstanceOwnerID; got != "amazon-rds" {
		t.Errorf("service owner should be preserved, got %q", got)
	}
}

func TestSanitizeIsIdempotent(t *testing.T) {
	s := &Snapshot{Account: RedactedAccount}
	Sanitize(s) // must not panic or alter an already-redacted snapshot
	if s.Account != RedactedAccount {
		t.Errorf("Account = %q", s.Account)
	}
}
