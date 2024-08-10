package application

import (
	"fmt"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	"github.com/gohugonet/hugoverse/pkg/testkit"
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

	config, err := configFact.LoadConfig()
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

	if got := config.Language.DefaultLanguageKey(); got != "zh" {
		t.Errorf("Expected database.user to be 'zh', but got '%s'", got)
	}
}

func TestModule(t *testing.T) {
	tmpDir, clean, err := testkit.MkTestModule()
	defer clean()

	if err != nil {
		t.Fatalf("MkTestConfig returned an error: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config, err := configFact.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	mods, err := moduleFact.New(config)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	if len(mods.All()) != 2 {
		t.Fatalf("Expected 2 modules, but got %d", len(mods.All()))
	}
}

func TestFs(t *testing.T) {
	tmpDir, clean, err := testkit.MkTestContent()
	defer clean()

	if err != nil {
		t.Fatalf("MkTestConfig returned an error: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config, err := configFact.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	mods, err := moduleFact.New(config)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	fsInstance, err := fsFact.New(config, mods)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	if fsInstance.Content == nil {
		t.Fatalf("Content should not be nil")
	}

	var files []fs.FileMetaInfo

	walk := func(path string, info fs.FileMetaInfo) error {
		fmt.Println("???", path)

		if info.IsDir() {
			return nil
		}
		files = append(files, info)

		return nil
	}

	if err := fsInstance.WalkContent("", fs.WalkCallback{
		HookPre:  nil,
		WalkFn:   walk,
		HookPost: nil,
	}, fs.WalkwayConfig{}); err != nil {
		t.Fatalf("WalkContent returned an error: %v", err)
	}

	if len(files) != 2 {
		t.Fatalf("Expected 2 modules, but got %d", len(files))
	}
}
