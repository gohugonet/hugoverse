package factory

import (
	"github.com/mdfriday/hugoverse/pkg/testkit"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmpDir, clean, err := testkit.MkTestConfig()
	defer clean()

	if err != nil {
		t.Fatalf("MkTestConfig returned an error: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}
	if config == nil {
		t.Fatalf("Config should not be nil")
	}

	if got := config.Provider.GetString("googleAnalytics"); got != "G-STPKPBQR5Y" {
		t.Errorf("Expected database.user to be 'G-STPKPBQR5Y', but got '%s'", got)
	}

	if got := config.Root.BaseURL; got != "https://hugo.notes.sunwei.xyz/" {
		t.Errorf("Expected database.user to be 'https://hugo.notes.sunwei.xyz/', but got '%s'", got)
	}

	if got := config.Language.DefaultLanguage(); got != "zh" {
		t.Errorf("Expected database.user to be 'zh', but got '%s'", got)
	}
}
