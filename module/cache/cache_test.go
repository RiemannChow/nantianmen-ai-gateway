package cache

import (
	"testing"
)

func TestNormalizeStable(t *testing.T) {
	a := normalize("  Hello World\n")
	b := normalize("hello world")
	if a != b {
		t.Fatalf("normalize not stable: %q vs %q", a, b)
	}
	if a != "hello world" {
		t.Fatalf("unexpected normalize output: %q", a)
	}
}

func TestHashPromptDeterministic(t *testing.T) {
	h1 := hashPrompt("same prompt")
	h2 := hashPrompt("same prompt")
	if h1 != h2 {
		t.Fatal("hash not deterministic")
	}
	if hashPrompt("other") == h1 {
		t.Fatal("different prompts collide")
	}
}

func TestBuildKeyPrefixAndShape(t *testing.T) {
	k := buildKey("t1", "gpt-4", "normalized")
	if len(k) <= len(keyPrefix) {
		t.Fatalf("key too short: %s", k)
	}
	// tenant/model are part of the key namespace
	if buildKey("t1", "gpt-4", "x") == buildKey("t2", "gpt-4", "x") {
		t.Fatal("different tenants should yield different keys")
	}
}
