package manifest_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DavDaz/llm-wiki-generator/internal/manifest"
)

func TestNew_defaults(t *testing.T) {
	m := manifest.New("Legal Wiki", "legal-wiki", "")
	assert.Equal(t, "Legal Wiki", m.Wiki.Name)
	assert.Equal(t, "legal-wiki", m.Wiki.Slug)
	assert.Equal(t, "es", m.Wiki.Language)
	assert.Equal(t, "1.0.0", m.Wiki.TemplateVersion)
	assert.True(t, m.Tools.ClaudeCode)
	assert.False(t, m.Tools.OpenCode)
	assert.False(t, m.Tools.Pi)
}

func TestSaveLoad_roundtrip(t *testing.T) {
	dir := t.TempDir()

	original := manifest.New("Test Wiki", "test-wiki", "en")
	original.Tools.OpenCode = true
	original.Domain.PrimaryEntities = []string{"usuario", "rol"}
	original.Domain.Conventions = []manifest.Convention{{Rule: "cite sources"}}

	require.NoError(t, original.Save(dir))

	loaded, err := manifest.Load(dir)
	require.NoError(t, err)

	assert.Equal(t, original.Wiki.Name, loaded.Wiki.Name)
	assert.Equal(t, original.Wiki.Slug, loaded.Wiki.Slug)
	assert.Equal(t, original.Wiki.Language, loaded.Wiki.Language)
	assert.True(t, loaded.Tools.ClaudeCode)
	assert.True(t, loaded.Tools.OpenCode)
	assert.False(t, loaded.Tools.Pi)
	assert.Equal(t, []string{"usuario", "rol"}, loaded.Domain.PrimaryEntities)
	assert.Equal(t, "cite sources", loaded.Domain.Conventions[0].Rule)
}

func TestLoad_missing_file(t *testing.T) {
	_, err := manifest.Load(t.TempDir())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "wiki.toml")
}

func TestLoad_invalid_toml(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, manifest.Filename), []byte("NOT_TOML:::"), 0o644))
	_, err := manifest.Load(dir)
	assert.Error(t, err)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*manifest.Manifest)
		wantErr string
	}{
		{
			name:    "valid manifest",
			mutate:  func(_ *manifest.Manifest) {},
			wantErr: "",
		},
		{
			name:    "empty name",
			mutate:  func(m *manifest.Manifest) { m.Wiki.Name = "" },
			wantErr: "wiki.name is required",
		},
		{
			name:    "invalid slug — spaces",
			mutate:  func(m *manifest.Manifest) { m.Wiki.Slug = "my wiki" },
			wantErr: "wiki.slug must be kebab-case",
		},
		{
			name:    "invalid slug — uppercase",
			mutate:  func(m *manifest.Manifest) { m.Wiki.Slug = "MyWiki" },
			wantErr: "wiki.slug must be kebab-case",
		},
		{
			name:    "empty language",
			mutate:  func(m *manifest.Manifest) { m.Wiki.Language = "" },
			wantErr: "wiki.language is required",
		},
		{
			name: "no tools enabled",
			mutate: func(m *manifest.Manifest) {
				m.Tools.ClaudeCode = false
				m.Tools.OpenCode = false
				m.Tools.Pi = false
			},
			wantErr: "at least one tool must be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := manifest.New("Test Wiki", "test-wiki", "es")
			tt.mutate(m)
			err := m.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestEnabledTools(t *testing.T) {
	m := manifest.New("Wiki", "wiki", "es")
	m.Tools.ClaudeCode = true
	m.Tools.OpenCode = true
	m.Tools.Pi = false

	tools := m.EnabledTools()
	assert.Equal(t, []string{"claude-code", "opencode"}, tools)
}
