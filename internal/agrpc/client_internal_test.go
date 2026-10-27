package agrpc

import "testing"

func TestDNSTarget(t *testing.T) {
	cases := []struct {
		name string
		addr string
		want string
	}{
		{
			name: "hostname gets the dns scheme so grpc re-resolves it",
			addr: "brain:50051",
			want: "dns:///brain:50051",
		},
		{
			name: "IPv4 literal is left untouched, nothing to re-resolve",
			addr: "127.0.0.1:50051",
			want: "127.0.0.1:50051",
		},
		{
			name: "already-scoped target is left untouched",
			addr: "dns:///brain:50051",
			want: "dns:///brain:50051",
		},
		{
			name: "passthrough-scoped target is left untouched",
			addr: "passthrough:///brain:50051",
			want: "passthrough:///brain:50051",
		},
		{
			name: "malformed address (no port) is returned as-is",
			addr: "brain",
			want: "brain",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dnsTarget(tc.addr)
			if got != tc.want {
				t.Errorf("dnsTarget(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}
