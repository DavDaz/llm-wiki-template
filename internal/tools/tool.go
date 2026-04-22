// Package tools defines the ToolSupport interface and implementations for each
// supported AI tool backend (Claude Code, OpenCode, Pi).
package tools

import (
	"github.com/DavDaz/llm-wiki-template/internal/manifest"
)

// ToolSupport defines the contract each AI tool backend must satisfy.
// Implementations know which files to create/delete for their tool.
type ToolSupport interface {
	// Name returns the canonical tool identifier (e.g. "claude-code").
	Name() string

	// IsInstalled reports whether the tool's files exist in wikiRoot.
	IsInstalled(wikiRoot string) bool

	// Install creates all files and directories needed to enable this tool.
	// It must be idempotent — calling it on an already-installed tool is safe.
	Install(wikiRoot string, m *manifest.Manifest) error

	// Uninstall removes the tool's exclusive files from wikiRoot.
	// It receives the manifest so it can check whether shared files (e.g.
	// AGENTS.md) are still needed by another enabled tool before deleting them.
	Uninstall(wikiRoot string, m *manifest.Manifest) error

	// SharedFiles returns paths (relative to wikiRoot) that this tool shares
	// with other tools. The caller must not delete these when other tools that
	// also declare them are still enabled.
	SharedFiles() []string
}
