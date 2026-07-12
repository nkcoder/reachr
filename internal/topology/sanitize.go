package topology

import "strings"

// RedactedAccount is the placeholder substituted for the real AWS account id.
const RedactedAccount = "000000000000"

// Sanitize redacts the AWS account id from a snapshot so it can be committed as a
// fixture or shared. It replaces the account everywhere it appears — the Account
// field, ARNs, SG-peer accounts, and ENI instance-owner ids — while keeping resource
// ids and RFC1918 private IPs (not sensitive, and needed for realistic reasoning).
func Sanitize(s *Snapshot) {
	account := s.Account
	if account == "" || account == RedactedAccount {
		return
	}
	redact := func(v string) string { return strings.ReplaceAll(v, account, RedactedAccount) }

	s.Account = RedactedAccount
	for i := range s.Subnets {
		s.Subnets[i].ARN = redact(s.Subnets[i].ARN)
	}
	for i := range s.SecurityGroups {
		s.SecurityGroups[i].ARN = redact(s.SecurityGroups[i].ARN)
		for j := range s.SecurityGroups[i].Rules {
			if p := &s.SecurityGroups[i].Rules[j].Peer; p.Account == account {
				p.Account = RedactedAccount
			}
		}
	}
	for i := range s.ENIs {
		if a := s.ENIs[i].Attachment; a != nil && a.InstanceOwnerID == account {
			a.InstanceOwnerID = RedactedAccount
		}
	}
}
