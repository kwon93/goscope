package capture

import "testing"

func TestValidateBPFFilter_Empty(t *testing.T) {
	if err := ValidateBPFFilter(""); err != nil {
		t.Fatalf("ValidateBPFFilter(\"\") = %v; want nil", err)
	}
}

func TestValidateBPFFilter_Valid(t *testing.T) {
	cases := []string{
		"tcp port 80",
		"udp port 53",
		"host 192.168.1.1",
		"tcp and port 443",
		"not port 22",
	}
	for _, filter := range cases {
		t.Run(filter, func(t *testing.T) {
			if err := ValidateBPFFilter(filter); err != nil {
				t.Fatalf("ValidateBPFFilter(%q) = %v; want nil", filter, err)
			}
		})
	}
}

func TestValidateBPFFilter_Invalid(t *testing.T) {
	cases := []string{
		"tcp 80",
		"port",
		"invalid@filter",
		"tcp port port",
	}
	for _, filter := range cases {
		t.Run(filter, func(t *testing.T) {
			if err := ValidateBPFFilter(filter); err == nil {
				t.Fatalf("ValidateBPFFilter(%q) = nil; want error", filter)
			}
		})
	}
}
