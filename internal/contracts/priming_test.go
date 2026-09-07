package contracts

import "testing"

func TestPrimingPreflightChecksWholeInventory(t *testing.T) {
	p := "memory-bank/flows/priming/test.yaml"
	good := map[string][]byte{p: []byte("version: 1\nprocess: test\nstages:\n  entry:\n    - memory-bank/dna/*.md\n"), "memory-bank/dna/README.md": []byte("index")}
	if err := ValidatePriming(good); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"memory-bank/missing.md", "memory-bank/../outside.md", "memory-bank/dna/**", "memory-bank/<bad>/x.md"} {
		bad := map[string][]byte{}
		for k, v := range good {
			bad[k] = v
		}
		bad[p] = []byte("version: 1\nprocess: test\nstages:\n  entry:\n    - " + input + "\n")
		if err := ValidatePriming(bad); err == nil {
			t.Fatalf("invalid input accepted: %s", input)
		}
	}
}
