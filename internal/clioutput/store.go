package clioutput

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CacheState struct {
	FilePath     string
	Document     *Document
	DetailFresh  bool
	ReviewsFresh bool
}

func (s CacheState) NeedsDetail() bool  { return !s.DetailFresh }
func (s CacheState) NeedsReviews() bool { return !s.ReviewsFresh }

type Store struct {
	outputDir string
	subDir    string
	now       func() time.Time
	// rename is overridable in tests to force an atomic-write failure.
	rename func(oldpath, newpath string) error
}

func NewStore(outputDir string, now func() time.Time) *Store {
	if now == nil {
		now = time.Now
	}
	return &Store{
		outputDir: outputDir,
		now:       now,
		rename:    os.Rename,
	}
}

// NewSubdirStore creates a Store that writes files into outputDir/<subDir>/
func NewSubdirStore(outputDir, subDir string, now func() time.Time) *Store {
	return &Store{
		outputDir: filepath.Join(outputDir, subDir),
		now:       now,
		rename:    os.Rename,
	}
}

func (s *Store) FilePath(id string) (string, error) {
	return FilePath(s.outputDir, id)
}

// FilePathWithCode returns the cache file path, preferring the video code
// (番号) over the internal ID for the filename.
func (s *Store) FilePathWithCode(code, id string) (string, error) {
	return FilePathWithCode(s.outputDir, code, id)
}

// Load reads the cache file for id. A document that is missing, corrupt, or
// not on the current SchemaVersion is treated as a full cache miss rather
// than migrated. Detail and reviews freshness are judged independently: a
// video whose reviews last failed can still report DetailFresh so only
// reviews are retried.
func (s *Store) Load(id string, staleAfter time.Duration) (CacheState, error) {
	filePath, err := s.FilePath(id)
	if err != nil {
		return CacheState{}, err
	}
	return s.loadAtPath(filePath, staleAfter)
}

// LoadByCode loads cache by code (番号) first, falling back to id.
func (s *Store) LoadByCode(code, id string, staleAfter time.Duration) (CacheState, error) {
	filePath, err := s.FilePathWithCode(code, id)
	if err != nil {
		return CacheState{}, err
	}
	return s.loadAtPath(filePath, staleAfter)
}

func (s *Store) loadAtPath(filePath string, staleAfter time.Duration) (CacheState, error) {
	state := CacheState{FilePath: filePath}

	body, err := os.ReadFile(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return CacheState{}, fmt.Errorf("read cache file: %w", err)
	}

	var doc Document
	if err := json.Unmarshal(body, &doc); err != nil {
		return state, nil
	}
	if doc.SchemaVersion != SchemaVersion {
		return state, nil
	}

	state.Document = &doc
	state.DetailFresh = s.now().Sub(doc.Metadata.DetailUpdatedAt) <= staleAfter
	if doc.Metadata.ReviewsUpdatedAt != nil {
		state.ReviewsFresh = s.now().Sub(*doc.Metadata.ReviewsUpdatedAt) <= staleAfter
	}
	return state, nil
}

// WriteFile atomically persists doc, keyed by doc.Detail.Summary.Code (番号)
// when available, falling back to Summary.ID. Sources are merged with any
// existing cache file's sources before serialization; a rename failure leaves
// the previous valid document untouched.
func (s *Store) WriteFile(doc Document) error {
	code := doc.Detail.Summary.Code
	id := string(doc.Detail.Summary.ID)
	filePath, err := s.FilePathWithCode(code, id)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(s.outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	doc, err = s.mergeSources(code, id, doc)
	if err != nil {
		return err
	}

	body, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}
	body = append(body, '\n')

	return s.atomicWrite(filePath, body)
}

func (s *Store) atomicWrite(filePath string, body []byte) error {
	tmp, err := os.CreateTemp(s.outputDir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := s.rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}
	return nil
}

func (s *Store) WriteJSON(w io.Writer, doc Document) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}

func MergeSources(existing []Source, next []Source) []Source {
	merged := make([]Source, 0, len(existing)+len(next))
	seen := make(map[string]struct{}, len(existing)+len(next))

	for _, source := range append(append([]Source{}, existing...), next...) {
		key := source.Key()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		merged = append(merged, source)
	}

	return merged
}

func (s *Store) mergeSources(code, id string, doc Document) (Document, error) {
	state, err := s.LoadByCode(code, id, 0)
	if err != nil {
		return Document{}, err
	}
	if state.Document == nil {
		return doc, nil
	}

	doc.Metadata.Sources = MergeSources(state.Document.Metadata.Sources, doc.Metadata.Sources)
	return doc, nil
}

// NFOPath returns the .nfo file path for the given video ID, placed in the
// same directory as the JSON cache file.
func (s *Store) NFOPath(id string) (string, error) {
	jsonPath, err := s.FilePath(id)
	if err != nil {
		return "", err
	}
	return jsonPath[:len(jsonPath)-len(".json")] + ".nfo", nil
}

// WriteNFO generates a Jellyfin/Plex-compatible .nfo file from the document.
func (s *Store) WriteNFO(path string, doc Document) error {
	d := doc.Detail
	var sb strings.Builder

	sb.WriteString("<musicvideo>\n")

	code := d.Summary.Code
	title := d.Summary.Title
	cleanName := code
	if cleanName == "" {
		cleanName = title
	}
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", escapeNFO(cleanName)))
	sb.WriteString(fmt.Sprintf("  <plot>%s</plot>\n", escapeNFO(title)))

	if !d.Summary.PublishedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("  <year>%s</year>\n", d.Summary.PublishedAt.Format("2006-01-02")))
	}

	if d.DurationMinutes != nil && *d.DurationMinutes > 0 {
		sb.WriteString(fmt.Sprintf("  <runtime>%d</runtime>\n", *d.DurationMinutes))
	}

	if d.Summary.Score.Value > 0 {
		sb.WriteString(fmt.Sprintf("  <rating>%.1f</rating>\n", d.Summary.Score.Value))
	}

	if d.Summary.Score.Count > 0 {
		sb.WriteString(fmt.Sprintf("  <votes>%d</votes>\n", d.Summary.Score.Count))
	}

	for _, tag := range d.Tags {
		sb.WriteString(fmt.Sprintf("  <tag>%s</tag>\n", escapeNFO(tag.Name)))
	}

	for _, actor := range d.Actors {
		sb.WriteString("  <actor>\n")
		sb.WriteString(fmt.Sprintf("    <name>%s</name>\n", escapeNFO(actor.Name)))
		sb.WriteString("    <type>actor</type>\n")
		sb.WriteString("  </actor>\n")
	}

	if d.Maker != nil && d.Maker.Name != "" {
		sb.WriteString(fmt.Sprintf("  <studio>%s</studio>\n", escapeNFO(d.Maker.Name)))
	}

	if d.Director != nil && d.Director.Name != "" {
		sb.WriteString(fmt.Sprintf("  <director>%s</director>\n", escapeNFO(d.Director.Name)))
	}

	if d.Summary.CoverURL != "" {
		sb.WriteString(fmt.Sprintf("  <thumb>%s</thumb>\n", d.Summary.CoverURL))
	}

	for _, m := range d.Magnets {
		sb.WriteString(fmt.Sprintf("  <source>%s</source>\n", escapeNFO(m.URI)))
	}

	sb.WriteString("</musicvideo>")

	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		dir := path[:idx]
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create nfo dir: %w", err)
		}
	} else if path != "" {
		if err := os.MkdirAll(".", 0o755); err != nil {
			return fmt.Errorf("create nfo dir: %w", err)
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}

func escapeNFO(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
