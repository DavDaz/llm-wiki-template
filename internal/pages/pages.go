// Package pages provides frontmatter-aware read/write helpers for wiki pages.
package pages

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Status represents the workflow status stored in page frontmatter.
type Status string

const (
	StatusVigente   Status = "vigente"
	StatusBorrador  Status = "borrador"
	StatusDeprecado Status = "deprecado"
)

// Page is surfaced metadata for a markdown page.
type Page struct {
	Title     string
	Type      string
	Status    Status
	UpdatedAt time.Time
	Path      string
}

// List scans wikiDir non-recursively and parses direct *.md files.
// Malformed files are skipped and counted instead of failing the whole operation.
func List(wikiDir string) ([]Page, int, error) {
	entries, err := os.ReadDir(wikiDir)
	if err != nil {
		return nil, 0, fmt.Errorf("read wiki dir %s: %w", wikiDir, err)
	}

	result := make([]Page, 0, len(entries))
	skipped := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		absPath, err := filepath.Abs(filepath.Join(wikiDir, entry.Name()))
		if err != nil {
			skipped++
			continue
		}

		buf, err := os.ReadFile(absPath)
		if err != nil {
			skipped++
			continue
		}

		page, ok := parsePage(absPath, buf)
		if !ok {
			skipped++
			continue
		}
		result = append(result, page)
	}

	return result, skipped, nil
}

// FilterByStatus returns pages with matching status, preserving order.
func FilterByStatus(ps []Page, s Status) []Page {
	filtered := make([]Page, 0)
	for _, p := range ps {
		if p.Status == s {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// SetStatus rewrites only the status frontmatter line using an atomic tmp+rename.
func SetStatus(path string, newStatus Status) error {
	if !isValidStatus(newStatus) {
		return fmt.Errorf("unknown status %q", newStatus)
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	statusLineStart, statusLineEnd, lineTerminator, err := findStatusLine(buf, path)
	if err != nil {
		return err
	}

	lineContent := buf[statusLineStart:statusLineEnd]
	indent := leadingWhitespace(lineContent)
	replacement := append([]byte(indent+"status: "+string(newStatus)), lineTerminator...)

	out := make([]byte, 0, len(buf)-len(buf[statusLineStart:statusLineEnd+len(lineTerminator)])+len(replacement))
	out = append(out, buf[:statusLineStart]...)
	out = append(out, replacement...)
	out = append(out, buf[statusLineEnd+len(lineTerminator):]...)

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, out, 0o644); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file %s to %s: %w", tmpPath, path, err)
	}

	return nil
}

func parsePage(absPath string, buf []byte) (Page, bool) {
	fmStart, openFenceLen, ok := locateOpeningFence(buf)
	if !ok {
		return Page{}, false
	}

	fmContentStart := fmStart + openFenceLen
	_, closeFenceStart, ok := locateClosingFence(buf, fmContentStart)
	if !ok {
		return Page{}, false
	}

	meta := map[string]string{}
	for _, line := range splitLines(buf[fmContentStart:closeFenceStart]) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			return Page{}, false
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, "\"'")
		meta[key] = value
	}

	statusRaw, ok := meta["status"]
	if !ok {
		return Page{}, false
	}

	return Page{
		Title:     meta["titulo"],
		Type:      meta["tipo"],
		Status:    Status(statusRaw),
		UpdatedAt: parseUpdatedAt(meta["actualizado"]),
		Path:      absPath,
	}, true
}

func parseUpdatedAt(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}

	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.UTC()
	}

	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC()
	}

	return time.Time{}
}

func findStatusLine(buf []byte, path string) (lineStart int, lineEnd int, terminator []byte, err error) {
	fmStart, openFenceLen, ok := locateOpeningFence(buf)
	if !ok {
		return 0, 0, nil, fmt.Errorf("no frontmatter at %s", path)
	}

	fmContentStart := fmStart + openFenceLen
	fmContentEnd, _, ok := locateClosingFence(buf, fmContentStart)
	if !ok {
		return 0, 0, nil, fmt.Errorf("unterminated frontmatter in %s", path)
	}

	for _, line := range iterateLines(buf, fmContentStart, fmContentEnd) {
		trimmed := strings.TrimSpace(string(line.content))
		if strings.HasPrefix(trimmed, "status:") {
			if strings.TrimSpace(strings.TrimPrefix(trimmed, "status:")) == "" {
				continue
			}
			return line.start, line.endWithoutTerm, line.terminator, nil
		}
	}

	return 0, 0, nil, fmt.Errorf("no status key in frontmatter: %s", path)
}

func locateOpeningFence(buf []byte) (start int, fenceLen int, ok bool) {
	start = 0
	if bytes.HasPrefix(buf, []byte{0xEF, 0xBB, 0xBF}) {
		start = 3
	}
	if bytes.HasPrefix(buf[start:], []byte("---\r\n")) {
		return start, len("---\r\n"), true
	}
	if bytes.HasPrefix(buf[start:], []byte("---\n")) {
		return start, len("---\n"), true
	}
	return 0, 0, false
}

func locateClosingFence(buf []byte, fmContentStart int) (contentEnd int, fenceStart int, ok bool) {
	for _, line := range iterateLines(buf, fmContentStart, len(buf)) {
		if string(line.content) == "---" {
			return line.start, line.start, true
		}
	}
	return 0, 0, false
}

type parsedLine struct {
	start          int
	endWithoutTerm int
	terminator     []byte
	content        []byte
}

func iterateLines(buf []byte, start int, end int) []parsedLine {
	lines := make([]parsedLine, 0)
	i := start
	for i < end {
		lineStart := i
		for i < end && buf[i] != '\n' && buf[i] != '\r' {
			i++
		}
		lineEndWithoutTerm := i

		term := []byte{}
		if i < end {
			if buf[i] == '\r' {
				if i+1 < end && buf[i+1] == '\n' {
					term = []byte("\r\n")
					i += 2
				} else {
					term = []byte("\r")
					i++
				}
			} else {
				term = []byte("\n")
				i++
			}
		}

		lines = append(lines, parsedLine{
			start:          lineStart,
			endWithoutTerm: lineEndWithoutTerm,
			terminator:     term,
			content:        buf[lineStart:lineEndWithoutTerm],
		})
	}
	return lines
}

func splitLines(buf []byte) []string {
	raw := strings.Split(strings.ReplaceAll(string(buf), "\r\n", "\n"), "\n")
	out := make([]string, 0, len(raw))
	for _, line := range raw {
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}

func leadingWhitespace(line []byte) string {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	return string(line[:i])
}

func isValidStatus(s Status) bool {
	return s == StatusVigente || s == StatusBorrador || s == StatusDeprecado
}
