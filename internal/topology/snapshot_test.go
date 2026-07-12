package topology

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func p(v int32) *int32 { return &v }

// TestRoundTrip guards schema stability and the pointer-port nil semantics
// (nil FromPort/ToPort = "all ports", which must survive marshal/unmarshal).
func TestRoundTrip(t *testing.T) {
	want := &Snapshot{
		SchemaVersion: SchemaVersion,
		ScannedAt:     time.Date(2026, 7, 12, 9, 58, 23, 0, time.UTC),
		Region:        "ap-southeast-2",
		Account:       RedactedAccount,
		RouteTables: []RouteTable{{
			ID: "rtb-1", VPCID: "vpc-1", Main: true,
			Routes: []Route{{
				Destination: RouteDestination{Kind: DestPrefixList, Value: "pl-1"},
				Target:      RouteTarget{Kind: TargetVPCEndpointGW, ID: "vpce-1"},
				State:       "active",
			}},
		}},
		SecurityGroups: []SecurityGroup{{
			ID: "sg-1", VPCID: "vpc-1", Name: "app",
			Rules: []SGRule{
				{Direction: Ingress, Protocol: "tcp", FromPort: p(8080), ToPort: p(8080), Peer: SGPeer{Kind: PeerSG, Value: "sg-2"}},
				{Direction: Ingress, Protocol: "all", Peer: SGPeer{Kind: PeerSG, Value: "sg-1"}}, // nil ports = all
			},
		}},
	}

	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got Snapshot
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(want, &got) {
		t.Errorf("round-trip mismatch:\n want %+v\n got  %+v", want, &got)
	}
	if got.SecurityGroups[0].Rules[1].FromPort != nil {
		t.Error("nil FromPort (all-ports) did not survive round-trip")
	}
}
