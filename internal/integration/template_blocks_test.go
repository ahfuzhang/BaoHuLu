package integration

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreservedTemplateBlocks(t *testing.T) {
	contents, err := os.ReadFile("testdata/template-blocks.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Baseline string `json:"baseline"`
		Blocks   []struct {
			Source      string `json:"source"`
			First       int    `json:"first"`
			Last        int    `json:"last"`
			Target      string `json:"target"`
			TargetFirst int    `json:"target_first"`
			Hash        string `json:"sha256"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Blocks) != 45 {
		t.Fatalf("expected 45 frozen blocks, got %d", len(manifest.Blocks))
	}
	for _, block := range manifest.Blocks {
		contents, err := os.ReadFile(filepath.Join("../..", block.Target))
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(string(contents), "\n")
		first, last := block.TargetFirst-1, block.TargetFirst+block.Last-block.First
		if first < 0 || last > len(lines) || first >= last {
			t.Fatalf("invalid block %+v", block)
		}
		actual := sha256.Sum256([]byte(strings.Join(lines[first:last], "\n") + "\n"))
		if hex.EncodeToString(actual[:]) != block.Hash {
			t.Errorf("D16: %s no longer matches %s:%d-%d at %s", block.Target, block.Source, block.First, block.Last, manifest.Baseline)
		}
	}
}
