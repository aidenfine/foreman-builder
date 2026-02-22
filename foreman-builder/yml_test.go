package foremanbuilder_test

import (
	"strings"
	"testing"

	foremanbuilder "github.com/aidenfine/foreman-builder/foreman-builder"
)

func TestParseConfig(t *testing.T) {
	yml := `
packages:
  - package 0
  - package 1
  - package 2
  - package 3
`

	cfg, err := foremanbuilder.ParseConfig(strings.NewReader(yml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Packages) != 4 {
		t.Fatalf("expected 4 packages, got %d", len(cfg.Packages))
	}

	if cfg.Packages[0] != "package 0" {
		t.Fatalf("expected package 0, got %s", cfg.Packages[0])
	}
	if cfg.Packages[len(cfg.Packages)-1] != "package 3" {
		t.Fatalf("expected package 3, got %s", cfg.Packages[0])

	}
}

// func TestParseConfig_InvalidYAML(t *testing.T) {
// 	yml := `::::::: this is not valid yaml`

// 	_, err := foremanbuilder.ParseConfig(strings.NewReader(yml))
// 	if err == nil {
// 		t.Fatal("expected error but got nil")
// 	}
// }
