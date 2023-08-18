package config

import "testing"

func TestParse(t *testing.T) {

	err := Parse("../columbus.conf")
	if err != nil {
		t.Fatalf("FAIL: %s\n", err)
	}

	if MongoURI == "" {
		t.Fatalf("FAIL: MongoURI is empty\n")
	}

	if Address != "127.0.0.1:8080" {
		t.Fatalf("FAIL: Invalid Address: %s\n", Address)
	}

	if len(TrustedProxies) != 1 {
		t.Fatalf("FAIL: Invalid TrustedProxies: %#v\n", TrustedProxies)
	}
}
