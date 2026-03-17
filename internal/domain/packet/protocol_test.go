package packet

import "testing"

func TestProtocolConstants(t *testing.T) {
	tests := []struct {
		name string
		got  Protocol
		want Protocol
	}{
		{"TCP", TCP, "tcp"},
		{"UDP", UDP, "udp"},
		{"OTHER", OTHER, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("protocol = %v; want %v", tt.got, tt.want)
			}
		})
	}
}

func TestProtocolString(t *testing.T) {
	tests := []struct {
		proto Protocol
		want  string
	}{
		{TCP, "tcp"},
		{UDP, "udp"},
		{OTHER, "other"},
	}

	for _, tt := range tests {
		t.Run(string(tt.proto), func(t *testing.T) {
			got := string(tt.proto)
			if got != tt.want {
				t.Fatalf("string(protocol) = %v; want %v", got, tt.want)
			}
		})
	}
}
