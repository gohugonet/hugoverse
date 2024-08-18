package application

import (
	"context"
	configFact "github.com/gohugonet/hugoverse/internal/domain/config/factory"
	contentHubFact "github.com/gohugonet/hugoverse/internal/domain/contenthub/factory"
	"github.com/gohugonet/hugoverse/internal/domain/fs"
	fsFact "github.com/gohugonet/hugoverse/internal/domain/fs/factory"
	mdFact "github.com/gohugonet/hugoverse/internal/domain/markdown/factory"
	moduleFact "github.com/gohugonet/hugoverse/internal/domain/module/factory"
	rsFact "github.com/gohugonet/hugoverse/internal/domain/resources/factory"
	siteFact "github.com/gohugonet/hugoverse/internal/domain/site/factory"
	tmplFact "github.com/gohugonet/hugoverse/internal/domain/template/factory"
	bp "github.com/gohugonet/hugoverse/pkg/bufferpool"
	"github.com/gohugonet/hugoverse/pkg/testkit"
	"github.com/spf13/cast"
	"os"
	"strings"
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
		if info.IsDir() {
			return nil
		}
		files = append(files, info)

		f, err := info.Open()
		if err != nil {
			return err
		}
		defer f.Close()

		return nil
	}

	if err := fsInstance.WalkContent("", fs.WalkCallback{
		HookPre:  nil,
		WalkFn:   walk,
		HookPost: nil,
	}, fs.WalkwayConfig{}); err != nil {
		t.Fatalf("WalkContent returned an error: %v", err)
	}

	if len(files) != 7 {
		t.Fatalf("Expected 2 modules, but got %d", len(files))
	}
}

func TestResource(t *testing.T) {
	tmpDir, clean, err := testkit.MkTestResource()
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

	ws := &resourcesWorkspaceProvider{
		Config: config,
		Fs:     fsInstance,
	}
	resources, err := rsFact.NewResources(ws)
	if err != nil {
		t.Fatalf("New resource returned an error: %v", err)
	}

	r, err := resources.GetResource("book.scss")
	if err != nil {
		t.Fatalf("Get resource `book.scss` returned an error: %v", err)
	}

	if r.TargetPath() != "/book.scss" {
		t.Fatalf("Expected target path `/book.scss`, but got %s", r.TargetPath())
	}

	r, err = resources.ToCSS(r, make(map[string]any))
	if err != nil {
		t.Fatalf("ToCSS resource `book.scss` returned an error: %v", err)
	}

	c, err := r.Content(context.Background())

	if !strings.Contains(cast.ToString(c), "html") {
		t.Fatalf("Expected result contains `html`, but got %s", c)
	}

	if r.TargetPath() != "/book.css" {
		t.Fatalf("Expected resource target path `/book.css`, but got %s", r.TargetPath())
	}

	r, err = resources.Minify(r)
	if err != nil {
		t.Fatalf("Minify resource `book.scss` returned an error: %v", err)
	}

	r, err = resources.Fingerprint(r, "")
	if err != nil {
		t.Fatalf("Fingerprint resource `book.scss` returned an error: %v", err)
	}

	d := r.Data()

	if !checkIntegrity(d) {
		t.Fatalf("Expected integrity, but got %s", d)
	}
}

func checkIntegrity(data any) bool {
	if m, ok := data.(map[string]any); ok {
		if integrity, exists := m["Integrity"]; exists && integrity != "" {
			return true
		}
	}
	return false
}

func TestTemplate(t *testing.T) {
	tmpDir, clean, err := testkit.MkTestTemplate()
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

	ch, err := contentHubFact.New(&chServices{
		Config: config,
		Fs:     fsInstance,
		Module: mods,
	})
	if err != nil {
		t.Fatalf("New content hub returned an error: %v", err)
	}

	ws := &resourcesWorkspaceProvider{
		Config: config,
		Fs:     fsInstance,
	}
	resources, err := rsFact.NewResources(ws)
	if err != nil {
		t.Fatalf("New resource returned an error: %v", err)
	}

	s := siteFact.New(&siteServices{
		Config:     config,
		Fs:         fsInstance,
		ContentHub: ch,
		Resources:  resources,
	})

	exec, err := tmplFact.New(fsInstance, &templateCustomizedFunctionsProvider{
		Markdown:   mdFact.NewMarkdown(),
		ContentHub: ch,
		Site:       s,
		Resources:  resources,
		Config:     config,
		Fs:         fsInstance,
	})

	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}

	tmpl, found, err := exec.LookupLayout([]string{"index.html"})
	if err != nil {
		t.Fatalf("LookupLayout returned an error: %v", err)
	}
	if !found {
		t.Fatalf("Template not found")
	}

	data := testkit.NewTemplateIndex("Index", "Content")
	renderBuffer := bp.GetBuffer()
	defer bp.PutBuffer(renderBuffer)

	err = exec.ExecuteWithContext(context.Background(), tmpl, renderBuffer, data)
	if err != nil {
		t.Fatalf("ExecuteWithContext returned an error: %v", err)
	}

	if !strings.Contains(renderBuffer.String(), "<body>Content</body>") {
		t.Fatalf("Expected result not contains `<body>Content</body>`, but got %s", renderBuffer.String())
	}
}
