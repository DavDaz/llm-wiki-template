// Package manifest handles reading and writing wiki.toml — the single source
// of truth for a wiki's configuration and enabled tool backends.
package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const Filename = "wiki.toml"

var slugRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Wiki holds top-level wiki metadata.
type Wiki struct {
	Name            string `toml:"name"`
	Slug            string `toml:"slug"`
	Language        string `toml:"language"`
	TemplateVersion string `toml:"template_version"`
	CreatedAt       string `toml:"created_at"`
}

// Tools declares which AI tool backends are enabled for this wiki.
type Tools struct {
	ClaudeCode bool `toml:"claude_code"`
	OpenCode   bool `toml:"opencode"`
	Pi         bool `toml:"pi"`
}

// Convention is a domain-specific business rule the AI must enforce.
type Convention struct {
	Rule string `toml:"rule"`
}

// Domain describes the semantic domain the wiki covers.
type Domain struct {
	PrimaryEntities []string     `toml:"primary_entities"`
	PageTypes       []string     `toml:"page_types"`
	Conventions     []Convention `toml:"conventions"`
}

// Manifest is the full wiki.toml structure.
type Manifest struct {
	Wiki   Wiki   `toml:"wiki"`
	Tools  Tools  `toml:"tools"`
	Domain Domain `toml:"domain"`
}

// New returns a Manifest with sensible defaults for a new wiki.
func New(name, slug, lang string) *Manifest {
	if lang == "" {
		lang = "es"
	}
	return &Manifest{
		Wiki: Wiki{
			Name:            name,
			Slug:            slug,
			Language:        lang,
			TemplateVersion: "1.0.0",
			CreatedAt:       time.Now().Format("2006-01-02"),
		},
		Tools: Tools{
			ClaudeCode: true,
		},
		Domain: Domain{
			PrimaryEntities: []string{},
			PageTypes:       []string{"proceso", "referencia", "entidad", "politica"},
			Conventions:     []Convention{},
		},
	}
}

// Load reads wiki.toml from wikiRoot and returns the parsed manifest.
func Load(wikiRoot string) (*Manifest, error) {
	path := filepath.Join(wikiRoot, Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m Manifest
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &m, nil
}

// Save writes the manifest to wiki.toml in wikiRoot.
func (m *Manifest) Save(wikiRoot string) error {
	path := filepath.Join(wikiRoot, Filename)
	data, err := toml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// Validate returns an error if the manifest is semantically invalid.
func (m *Manifest) Validate() error {
	var errs []string

	if strings.TrimSpace(m.Wiki.Name) == "" {
		errs = append(errs, "wiki.name is required")
	}
	if !slugRe.MatchString(m.Wiki.Slug) {
		errs = append(errs, "wiki.slug must be kebab-case (e.g. my-legal-wiki)")
	}
	if m.Wiki.Language == "" {
		errs = append(errs, "wiki.language is required")
	}
	if !m.Tools.ClaudeCode && !m.Tools.OpenCode && !m.Tools.Pi {
		errs = append(errs, "at least one tool must be enabled (tools.claude_code, tools.opencode, or tools.pi)")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// EnabledTools returns a slice of tool names that are currently enabled.
func (m *Manifest) EnabledTools() []string {
	var tools []string
	if m.Tools.ClaudeCode {
		tools = append(tools, "claude-code")
	}
	if m.Tools.OpenCode {
		tools = append(tools, "opencode")
	}
	if m.Tools.Pi {
		tools = append(tools, "pi")
	}
	return tools
}
