package wol

import (
	"bytes"
	"net"
	"testing"
)

func TestMACValidationAndNormalization(t *testing.T) {
	for _, input := range []string{"AA:BB:CC:DD:EE:FF", "aa-bb-cc-dd-ee-ff", "aabb.ccdd.eeff", " aa:bb:cc:dd:ee:ff "} {
		if err := ValidateMAC(input); err != nil {
			t.Errorf("%q: %v", input, err)
		}
		got, err := NormalizeMAC(input)
		if err != nil || got != "aa:bb:cc:dd:ee:ff" {
			t.Errorf("%q: %q, %v", input, got, err)
		}
	}
	for _, input := range []string{"", "00:11:22:33:44", "00:11:22:33:44:55:66:77", "gg:11:22:33:44:55"} {
		if err := ValidateMAC(input); err == nil {
			t.Errorf("accepted invalid MAC %q", input)
		}
		if _, err := NormalizeMAC(input); err == nil {
			t.Errorf("normalized invalid MAC %q", input)
		}
		if ok, _ := SendMagicPacket(input, ""); ok {
			t.Errorf("sent invalid MAC %q", input)
		}
	}
}

func TestMagicPacketBytes(t *testing.T) {
	mac := net.HardwareAddr{0, 1, 2, 3, 4, 5}
	want := append(bytes.Repeat([]byte{255}, 6), bytes.Repeat(mac, 16)...)
	if got := createMagicPacket(mac); !bytes.Equal(got, want) {
		t.Fatalf("unexpected packet: %x", got)
	}
}

func network(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	ip, subnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatal(err)
	}
	subnet.IP = ip
	return subnet
}

func TestTargetsUseActualSubnetMasks(t *testing.T) {
	networks := []*net.IPNet{network(t, "10.10.5.55/16"), network(t, "192.168.1.130/25"), network(t, "172.16.0.1/31"), network(t, "10.0.0.1/32")}
	for _, tc := range []struct {
		ip   string
		want []string
	}{
		{"10.10.7.3", []string{"<nil>/10.10.7.3", "10.10.5.55/10.10.255.255", "<nil>/255.255.255.255"}},
		{"192.168.1.140", []string{"<nil>/192.168.1.140", "192.168.1.130/192.168.1.255", "<nil>/255.255.255.255"}},
		{"192.168.90.10", []string{"<nil>/192.168.90.10", "<nil>/255.255.255.255"}},
		{"", []string{"10.10.5.55/10.10.255.255", "192.168.1.130/192.168.1.255", "<nil>/255.255.255.255"}},
	} {
		t.Run(tc.ip, func(t *testing.T) {
			got := packetTargets(net.ParseIP(tc.ip), networks)
			if len(got) != len(tc.want) {
				t.Fatalf("got %d targets, want %d", len(got), len(tc.want))
			}
			for i, target := range got {
				if value := target.source.String() + "/" + target.destination.String(); value != tc.want[i] {
					t.Errorf("got %s, want %s", value, tc.want[i])
				}
			}
		})
	}
}

func TestInvalidIPRejectedBeforeSending(t *testing.T) {
	for _, ip := range []string{"bad", "::1", "192.168.1.256"} {
		if ok, _ := SendMagicPacket("00:11:22:33:44:55", ip); ok {
			t.Errorf("accepted %q", ip)
		}
	}
}
