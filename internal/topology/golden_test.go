package topology

import (
	"encoding/json"
	"os"
	"testing"
)

func loadGolden(t *testing.T) *Snapshot {
	t.Helper()
	b, err := os.ReadFile("../../testdata/testenv.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	var s Snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatalf("golden fixture does not parse: %v", err)
	}
	return &s
}

func TestGoldenCounts(t *testing.T) {
	s := loadGolden(t)
	for _, c := range []struct {
		name string
		got  int
		want int
	}{
		{"vpcs", len(s.VPCs), 1},
		{"subnets", len(s.Subnets), 2},
		{"routeTables", len(s.RouteTables), 2},
		{"enis", len(s.ENIs), 5},
		{"securityGroups", len(s.SecurityGroups), 4},
		{"vpcEndpoints", len(s.VPCEndpoints), 1},
	} {
		if c.got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, c.got, c.want)
		}
	}
	if s.SchemaVersion != SchemaVersion {
		t.Errorf("schemaVersion = %d, want %d", s.SchemaVersion, SchemaVersion)
	}
	if s.Account != RedactedAccount {
		t.Errorf("fixture account not sanitized: %q", s.Account)
	}
}

// TestGoldenSGChain asserts the alb -> app -> db SG-references-SG chain resolves.
func TestGoldenSGChain(t *testing.T) {
	s := loadGolden(t)
	byName := map[string]SecurityGroup{}
	for _, sg := range s.SecurityGroups {
		byName[sg.Name] = sg
	}
	alb, app, db := byName["reachr-testenv-alb"], byName["reachr-testenv-app"], byName["reachr-testenv-db"]
	if alb.ID == "" || app.ID == "" || db.ID == "" {
		t.Fatalf("missing a tier SG: alb=%q app=%q db=%q", alb.ID, app.ID, db.ID)
	}
	if !hasIngressFromSG(app, alb.ID) {
		t.Error("app SG should allow ingress from alb SG")
	}
	if !hasIngressFromSG(db, app.ID) {
		t.Error("db SG should allow ingress from app SG")
	}
}

func hasIngressFromSG(sg SecurityGroup, peerSGID string) bool {
	for _, r := range sg.Rules {
		if r.Direction == Ingress && r.Peer.Kind == PeerSG && r.Peer.Value == peerSGID {
			return true
		}
	}
	return false
}

// TestGoldenGatewayEndpointRoute asserts the S3 gateway-endpoint prefix-list route.
func TestGoldenGatewayEndpointRoute(t *testing.T) {
	s := loadGolden(t)
	for _, rt := range s.RouteTables {
		for _, r := range rt.Routes {
			if r.Target.Kind == TargetVPCEndpointGW && r.Destination.Kind == DestPrefixList {
				return
			}
		}
	}
	t.Error("expected a prefix-list -> vpc-endpoint-gw route in the fixture")
}

// TestGoldenENIOwners asserts the service-managed ENI owners survived.
func TestGoldenENIOwners(t *testing.T) {
	s := loadGolden(t)
	owners := map[string]bool{}
	for _, e := range s.ENIs {
		if e.Attachment != nil {
			owners[e.Attachment.InstanceOwnerID] = true
		}
	}
	for _, want := range []string{"amazon-elb", "amazon-rds"} {
		if !owners[want] {
			t.Errorf("expected an ENI owned by %q", want)
		}
	}
}
