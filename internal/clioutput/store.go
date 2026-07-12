package clioutput

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
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

func (s *Store) FilePath(id string) (string, error) {
	return FilePath(s.outputDir, id)
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

// WriteFile atomically persists doc, keyed by doc.Detail.Summary.ID. Sources
// are merged with any existing cache file's sources before serialization; a
// rename failure leaves the previous valid document untouched.
func (s *Store) WriteFile(doc Document) error {
	id := string(doc.Detail.Summary.ID)
	filePath, err := s.FilePath(id)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(s.outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	doc, err = s.mergeSources(id, doc)
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

func (s *Store) mergeSources(id string, doc Document) (Document, error) {
	state, err := s.Load(id, 0)
	if err != nil {
		return Document{}, err
	}
	if state.Document == nil {
		return doc, nil
	}

	doc.Metadata.Sources = MergeSources(state.Document.Metadata.Sources, doc.Metadata.Sources)
	return doc, nil
}
