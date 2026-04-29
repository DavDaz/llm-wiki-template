// Package generator orchestrates the creation and migration of wiki directories.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
	"github.com/DavDaz/llm-wiki-generator/internal/templates"
	"github.com/DavDaz/llm-wiki-generator/internal/tools"
)

// InitConfig holds the parameters for creating a new wiki.
type InitConfig struct {
	// ParentDir is the directory where the wiki folder will be created.
	// Defaults to the current working directory if empty.
	ParentDir string

	Name     string
	Slug     string
	Language string

	ClaudeCode bool
	OpenCode   bool
	Pi         bool

	PrimaryEntities []string
	PageTypes       []string
	Conventions     []string
}

// Init creates a new wiki directory from the given config.
// The wiki is created at ParentDir/<slug>-wiki/.
func Init(cfg InitConfig) (wikiRoot string, err error) {
	if cfg.ParentDir == "" {
		cfg.ParentDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("get working dir: %w", err)
		}
	}

	wikiRoot = filepath.Join(cfg.ParentDir, cfg.Slug)
	if _, err := os.Stat(wikiRoot); err == nil {
		return "", fmt.Errorf("directory already exists: %s", wikiRoot)
	}

	// Build manifest.
	m := manifest.New(cfg.Name, cfg.Slug, cfg.Language)
	m.Tools.ClaudeCode = cfg.ClaudeCode
	m.Tools.OpenCode = cfg.OpenCode
	m.Tools.Pi = cfg.Pi

	for _, e := range cfg.PrimaryEntities {
		m.Domain.PrimaryEntities = append(m.Domain.PrimaryEntities, e)
	}
	if len(cfg.PageTypes) > 0 {
		m.Domain.PageTypes = cfg.PageTypes
	}
	for _, c := range cfg.Conventions {
		m.Domain.Conventions = append(m.Domain.Conventions, manifest.Convention{Rule: c})
	}

	if err := m.Validate(); err != nil {
		return "", fmt.Errorf("invalid config: %w", err)
	}

	// Create directory structure.
	for _, dir := range []string{wikiRoot, filepath.Join(wikiRoot, "raw"), filepath.Join(wikiRoot, "wiki")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create dir %s: %w", dir, err)
		}
	}

	// Write manifest.
	if err := m.Save(wikiRoot); err != nil {
		return "", err
	}

	// Write wiki/index.md, wiki/log.md, and wiki/sources.json.
	if err := writeFile(filepath.Join(wikiRoot, "wiki", "index.md"), templates.RenderIndex(cfg.Name)); err != nil {
		return "", err
	}
	if err := writeFile(filepath.Join(wikiRoot, "wiki", "log.md"),
		templates.RenderLog(cfg.Name, cfg.Slug, time.Now().Format("2006-01-02"), cfg.PrimaryEntities, m.Domain.PageTypes)); err != nil {
		return "", err
	}
	if err := writeFile(filepath.Join(wikiRoot, "wiki", "sources.json"), templates.RenderSourcesRegistry()); err != nil {
		return "", err
	}

	// Write .gitignore.
	if err := writeFile(filepath.Join(wikiRoot, ".gitignore"), ".DS_Store\n*.swp\n*.tmp\n"); err != nil {
		return "", err
	}

	// Install enabled tools.
	if err := applyTools(wikiRoot, m); err != nil {
		return "", err
	}

	return wikiRoot, nil
}

// Migrate applies the current manifest state to the filesystem, enabling or
// disabling tools as needed and re-rendering instruction files.
// Call this after modifying m.Tools or m.Domain and saving the manifest.
func Migrate(wikiRoot string, m *manifest.Manifest) error {
	if err := m.Validate(); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	enabledMap := map[string]bool{}
	for _, name := range m.EnabledTools() {
		enabledMap[name] = true
	}

	for _, tool := range tools.All() {
		shouldBeInstalled := enabledMap[tool.Name()]
		isInstalled := tool.IsInstalled(wikiRoot)

		switch {
		case shouldBeInstalled && !isInstalled:
			if err := tool.Install(wikiRoot, m); err != nil {
				return fmt.Errorf("install %s: %w", tool.Name(), err)
			}
		case !shouldBeInstalled && isInstalled:
			if err := tool.Uninstall(wikiRoot, m); err != nil {
				return fmt.Errorf("uninstall %s: %w", tool.Name(), err)
			}
		case shouldBeInstalled && isInstalled:
			// Re-render instruction files in case domain config changed.
			if err := tool.Install(wikiRoot, m); err != nil {
				return fmt.Errorf("re-render %s: %w", tool.Name(), err)
			}
		}
	}
	return nil
}

// applyTools installs all tools that are enabled in the manifest.
func applyTools(wikiRoot string, m *manifest.Manifest) error {
	for _, name := range m.EnabledTools() {
		tool, err := tools.Get(name)
		if err != nil {
			return err
		}
		if err := tool.Install(wikiRoot, m); err != nil {
			return fmt.Errorf("install %s: %w", name, err)
		}
	}
	return nil
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
