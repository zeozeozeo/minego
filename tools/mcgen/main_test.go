package main

import "testing"

func TestKnownProtocol(t *testing.T) {
	for _, test := range []struct {
		version  string
		protocol int32
	}{
		{"1.21.11", 774},
		{"26.1", 775},
		{"26.2", 776},
	} {
		got, ok := knownProtocol(test.version)
		if !ok || got != test.protocol {
			t.Fatalf("knownProtocol(%q) = %d, %t; want %d, true", test.version, got, ok, test.protocol)
		}
	}
	if _, ok := knownProtocol("not-a-release"); ok {
		t.Fatal("unknown release accepted")
	}
}

func TestSplitVersions(t *testing.T) {
	got := splitVersions(" 26.1,1.21.11,26.1 ")
	if len(got) != 2 || got[0] != "26.1" || got[1] != "1.21.11" {
		t.Fatalf("splitVersions = %#v", got)
	}
}
