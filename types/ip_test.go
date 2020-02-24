package types_test

import (
	"testing"
	"github.com/gathering/gondulapi/types"
)

func TestIP(t *testing.T) {
	btos := []struct{ls string; s string}{
		{"192.168.0.1", "192.168.0.1"},
		{"::1", "::1"},
		{"fe80::77d6:6a51:13d6:b1ef/64", "fe80::77d6:6a51:13d6:b1ef/64"},
		{"fe80:0000::77d6:6a51:13d6:b1ef/64","fe80::77d6:6a51:13d6:b1ef/64"},
		{"db8:2001:0001::1","db8:2001:1::1"},
	}
	for _,item := range btos {
		i := types.IP{}
		want := item.s
		src := item.ls
		err := i.UnmarshalText([]byte(src))	
		if err != nil {
			t.Errorf("Failed to UnmarshalText for %s, error: %v", src, err)
			continue
		}
		got := i.String()
		if got != want {
			t.Errorf("String to ip to string for %s did not yield expected result. Wanted %s, got %s", src, want, got)
		}
	}
}
