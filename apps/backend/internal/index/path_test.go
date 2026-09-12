package index

import "testing"

func TestJoinPath(t *testing.T) {
	cases := []struct {
		parent, name, want string
	}{
		{"", "Kerjaan", "/Kerjaan"},
		{"/", "Kerjaan", "/Kerjaan"},
		{"/Kerjaan", "Sub", "/Kerjaan/Sub"},
		{"/Kerjaan/Sub", "laporan.pdf", "/Kerjaan/Sub/laporan.pdf"},
		{"/Kerjaan/", "Sub", "/Kerjaan/Sub"},
		{"", "/leading-slash", "/leading-slash"},
	}
	for _, tc := range cases {
		if got := JoinPath(tc.parent, tc.name); got != tc.want {
			t.Errorf("JoinPath(%q, %q) = %q, mau %q", tc.parent, tc.name, got, tc.want)
		}
	}
}
