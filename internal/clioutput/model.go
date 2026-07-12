package clioutput

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"time"

	javdbapi "github.com/RPbro/javdbapi"
)

// SchemaVersion is the fixed cache document version. v1 documents (missing
// or mismatched schema_version) are always treated as a cache miss; there is
// no migration path.
const SchemaVersion = 2

// PartialError records a component that failed while the rest of a Document
// still persisted successfully. In v1 only "reviews" is a valid Component.
// Kind is a stable machine code; Message must never contain a full request
// URL, search keyword, or response body.
type PartialError struct {
	Component string `json:"component"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
}

type Metadata struct {
	LastUpdated time.Time `json:"last_updated"`
	// DetailUpdatedAt and ReviewsUpdatedAt track freshness independently:
	// a failed reviews fetch never invalidates an already-cached Detail.
	DetailUpdatedAt        time.Time  `json:"detail_updated_at"`
	ReviewsUpdatedAt       *time.Time `json:"reviews_updated_at,omitempty"`
	ReviewsLastAttemptedAt *time.Time `json:"reviews_last_attempted_at,omitempty"`
	Sources                []Source   `json:"sources"`
}

// MarshalJSON guards Sources so it always serializes as "[]" rather than
// "null", matching the SDK's own slice convention.
func (m Metadata) MarshalJSON() ([]byte, error) {
	type alias Metadata
	out := alias(m)
	if out.Sources == nil {
		out.Sources = []Source{}
	}
	return json.Marshal(out)
}

type Document struct {
	SchemaVersion int                  `json:"schema_version"`
	Metadata      Metadata             `json:"metadata"`
	Detail        javdbapi.VideoDetail `json:"detail"`
	Reviews       []javdbapi.Review    `json:"reviews"`
	PartialErrors []PartialError       `json:"partial_errors"`
}

// MarshalJSON guards Reviews and PartialErrors so they always serialize as
// "[]" rather than "null", matching the SDK's own slice convention.
func (d Document) MarshalJSON() ([]byte, error) {
	type alias Document
	out := alias(d)
	if out.Reviews == nil {
		out.Reviews = []javdbapi.Review{}
	}
	if out.PartialErrors == nil {
		out.PartialErrors = []PartialError{}
	}
	return json.Marshal(out)
}

type Source struct {
	Command string          `json:"command"`
	Query   json.RawMessage `json:"query"`
}

func (s Source) Key() string {
	var compact bytes.Buffer
	if err := json.Compact(&compact, s.Query); err != nil {
		return s.Command + ":" + string(s.Query)
	}
	return s.Command + ":" + compact.String()
}

type searchSourceQuery struct {
	Keyword string `json:"keyword"`
	Page    int    `json:"page"`
}

type homeSourceQuery struct {
	Type   string `json:"type,omitempty"`
	Filter string `json:"filter,omitempty"`
	Sort   string `json:"sort,omitempty"`
	Page   int    `json:"page"`
}

type makerSourceQuery struct {
	ID     string `json:"id"`
	Filter string `json:"filter,omitempty"`
	Page   int    `json:"page"`
}

type actorSourceQuery struct {
	ID     string   `json:"id"`
	Filter []string `json:"filter,omitempty"`
	Page   int      `json:"page"`
}

type rankingSourceQuery struct {
	Period string `json:"period"`
	Type   string `json:"type"`
	Page   int    `json:"page"`
}

type videoSourceQuery struct {
	ID string `json:"id"`
}

func NewHomeSource(typ string, filter string, sort string, page int) Source {
	return mustMarshalSource("home", homeSourceQuery{
		Type:   typ,
		Filter: filter,
		Sort:   sort,
		Page:   page,
	})
}

func NewMakerSource(id string, filter string, page int) Source {
	return mustMarshalSource("maker", makerSourceQuery{
		ID:     id,
		Filter: filter,
		Page:   page,
	})
}

func NewSearchSource(keyword string, page int) Source {
	return mustMarshalSource("search", searchSourceQuery{
		Keyword: keyword,
		Page:    page,
	})
}

func NewActorSource(id string, filters []string, page int) Source {
	cleaned := make([]string, 0, len(filters))
	seen := make(map[string]struct{}, len(filters))
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue
		}
		if _, ok := seen[filter]; ok {
			continue
		}
		seen[filter] = struct{}{}
		cleaned = append(cleaned, filter)
	}
	slices.Sort(cleaned)
	return mustMarshalSource("actor", actorSourceQuery{
		ID:     id,
		Filter: cleaned,
		Page:   page,
	})
}

func NewRankingSource(period string, typ string, page int) Source {
	return mustMarshalSource("ranking", rankingSourceQuery{
		Period: period,
		Type:   typ,
		Page:   page,
	})
}

func NewVideoSource(id string) Source {
	return mustMarshalSource("video", videoSourceQuery{ID: id})
}

func mustMarshalSource(command string, query any) Source {
	raw, err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	return Source{Command: command, Query: raw}
}
