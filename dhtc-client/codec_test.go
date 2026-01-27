package dhtc_client

import (
	"net"
	"testing"
)

func TestUnmarshalCompactNodeInfos(t *testing.T) {
	// 20 bytes ID + 4 bytes IP + 2 bytes Port = 26 bytes
	ipv4Node := append(make([]byte, 20), 127, 0, 0, 1, 0, 80)
	// 20 bytes ID + 16 bytes IP + 2 bytes Port = 38 bytes
	ipv6Node := append(make([]byte, 20), net.ParseIP("::1")...)
	ipv6Node = append(ipv6Node, 0, 80)

	tests := []struct {
		name    string
		input   []byte
		wantLen int
		wantErr bool
	}{
		{"empty", []byte{}, 0, false},
		{"single ipv4", ipv4Node, 1, false},
		{"multiple ipv4", append(ipv4Node, ipv4Node...), 2, false},
		{"single ipv6", ipv6Node, 1, false},
		{"multiple ipv6", append(ipv6Node, ipv6Node...), 2, false},
		{"invalid length", append(ipv4Node, 1, 2, 3), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalCompactNodeInfos(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCompactNodeInfos() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("UnmarshalCompactNodeInfos() len = %v, want %v", len(got), tt.wantLen)
			}
			if len(got) > 0 {
				if tt.name == "single ipv4" {
					if !got[0].Addr.IP.Equal(net.IPv4(127, 0, 0, 1)) {
						t.Errorf("expected 127.0.0.1, got %v", got[0].Addr.IP)
					}
					if got[0].Addr.Port != 80 {
						t.Errorf("expected port 80, got %v", got[0].Addr.Port)
					}
				}
				if tt.name == "single ipv6" {
					if !got[0].Addr.IP.Equal(net.ParseIP("::1")) {
						t.Errorf("expected ::1, got %v", got[0].Addr.IP)
					}
					if got[0].Addr.Port != 80 {
						t.Errorf("expected port 80, got %v", got[0].Addr.Port)
					}
				}
			}
		})
	}
}

func TestCompactNodeInfo_MarshalBinary(t *testing.T) {
	node4 := CompactNodeInfo{
		ID: make([]byte, 20),
		Addr: net.UDPAddr{
			IP:   net.ParseIP("1.2.3.4"),
			Port: 1234,
		},
	}
	b4 := node4.MarshalBinary()
	if len(b4) != 26 {
		t.Errorf("expected 26 bytes for IPv4, got %d", len(b4))
	}

	node6 := CompactNodeInfo{
		ID: make([]byte, 20),
		Addr: net.UDPAddr{
			IP:   net.ParseIP("2001:db8::1"),
			Port: 1234,
		},
	}
	b6 := node6.MarshalBinary()
	if len(b6) != 38 {
		t.Errorf("expected 38 bytes for IPv6, got %d", len(b6))
	}
}

func TestUnmarshalCompactPeers(t *testing.T) {
	ipv4Peer := []byte{127, 0, 0, 1, 0, 80}
	ipv6Peer := append(net.ParseIP("::1"), 0, 80)

	tests := []struct {
		name    string
		input   []byte
		wantLen int
		wantErr bool
	}{
		{"empty", []byte{}, 0, false},
		{"single ipv4", ipv4Peer, 1, false},
		// 18 bytes will be interpreted as 3 IPv4 peers because it's multiple of 6.
		{"single ipv6 (interprets as 3 ipv4)", ipv6Peer, 3, false},
		{"multiple ipv4", append(ipv4Peer, ipv4Peer...), 2, false},
		{"invalid length", append(ipv4Peer, 1, 2, 3), 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalCompactPeers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCompactPeers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("UnmarshalCompactPeers() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestCompactPeer_UnmarshalBinary(t *testing.T) {
	ipv4Peer := []byte{127, 0, 0, 1, 0, 80}
	ipv6Peer := append(net.ParseIP("::1"), 0, 80)

	var cp4 CompactPeer
	if err := cp4.UnmarshalBinary(ipv4Peer); err != nil {
		t.Errorf("ipv4 UnmarshalBinary error: %v", err)
	}
	if !cp4.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Errorf("expected 127.0.0.1, got %v", cp4.IP)
	}

	var cp6 CompactPeer
	if err := cp6.UnmarshalBinary(ipv6Peer); err != nil {
		t.Errorf("ipv6 UnmarshalBinary error: %v", err)
	}
	if !cp6.IP.Equal(net.ParseIP("::1")) {
		t.Errorf("expected ::1, got %v", cp6.IP)
	}
}

func TestError_Bencode(t *testing.T) {
	e := Error{
		Code:    201,
		Message: []byte("Generic Error"),
	}
	b, err := e.MarshalBencode()
	if err != nil {
		t.Fatalf("MarshalBencode() error = %v", err)
	}

	var e2 Error
	err = e2.UnmarshalBencode(b)
	if err != nil {
		t.Fatalf("UnmarshalBencode() error = %v", err)
	}

	if e2.Code != e.Code {
		t.Errorf("expected code %d, got %d", e.Code, e2.Code)
	}
	if string(e2.Message) != string(e.Message) {
		t.Errorf("expected message %s, got %s", e.Message, e2.Message)
	}
}

func TestCompactPeers_Bencode(t *testing.T) {
	cps := CompactPeers{
		{IP: net.ParseIP("1.2.3.4"), Port: 1234},
		{IP: net.ParseIP("2001:db8::1"), Port: 5678},
	}

	b, err := cps.MarshalBencode()
	if err != nil {
		t.Fatalf("MarshalBencode() error = %v", err)
	}

	var cps2 CompactPeers
	if err := cps2.UnmarshalBencode(b); err != nil {
		t.Fatalf("UnmarshalBencode() error = %v", err)
	}

	if len(cps2) != len(cps) {
		t.Errorf("expected length %d, got %d", len(cps), len(cps2))
	}

	// Test unmarshal from single string (IPv4 only to avoid ambiguity)
	cps4 := CompactPeers{
		{IP: net.ParseIP("1.2.3.4"), Port: 1234},
		{IP: net.ParseIP("5.6.7.8"), Port: 5678},
	}
	singleString := append(cps4[0].MarshalBinary(), cps4[1].MarshalBinary()...)
	// Manually bencode a string
	bString := append([]byte("12:"), singleString...)
	var cps5 CompactPeers
	if err := cps5.UnmarshalBencode(bString); err != nil {
		t.Fatalf("UnmarshalBencode from string error = %v", err)
	}
	if len(cps5) != 2 {
		t.Errorf("expected length 2, got %d", len(cps5))
	}
}
