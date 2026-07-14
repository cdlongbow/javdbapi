package cliapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/RPbro/javdbapi"
	"github.com/RPbro/javdbapi/internal/clioutput"
	"github.com/RPbro/javdbapi/internal/siteurl"
)

// RunActorDetailCommand fetches the actor detail page and outputs JSON with
// id, name, and aliases.
func RunActorDetailCommand(ctx context.Context, store *clioutput.Store, req ListRequest) (Summary, error) {
	if err := validateSharedOptions(req.Shared); err != nil {
		return Summary{}, err
	}
	if req.ActorDetail == nil || req.ActorDetail.ActorID == "" {
		return Summary{}, fmt.Errorf("actor id is required")
	}

	actorID := string(req.ActorDetail.ActorID)

	client, err := javdbapi.NewClient(javdbapi.ClientConfig{})
	if err != nil {
		return Summary{}, err
	}

	target, err := siteurl.ActorDetail(client.BaseURL(), actorID, "zh")
	if err != nil {
		return Summary{}, err
	}

	body, err := client.GetRaw(ctx, target)
	if err != nil {
		return Summary{}, err
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return Summary{}, err
	}

	// Extract actor name
	var name string
	doc.Find(".actor-section-name").Each(func(i int, s *goquery.Selection) {
		if name == "" {
			name = strings.TrimSpace(s.Text())
		}
	})

	// Extract aliases (all section-meta spans except the one with video count)
	var aliases []string
	doc.Find(".section-meta").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text == "" || strings.Contains(text, "部影片") {
			return
		}
		parts := strings.Split(text, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				aliases = append(aliases, p)
			}
		}
	})

	if name == "" {
		return Summary{}, fmt.Errorf("could not find actor name")
	}

	result := map[string]any{
		"id":      actorID,
		"name":    name,
		"aliases": aliases,
	}

	enc := json.NewEncoder(req.Shared.Stdout)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(result); err != nil {
		return Summary{}, err
	}

	return Summary{Fetched: 1}, nil
}

// ActorNameFromHTML extracts the first actor name from the given HTML body.
func ActorNameFromHTML(body []byte) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	var name string
	doc.Find(".actor-section-name").Each(func(i int, s *goquery.Selection) {
		if name == "" {
			name = strings.TrimSpace(s.Text())
		}
	})

	if name == "" {
		return "", fmt.Errorf("could not find actor name")
	}

	// Split by comma and return the first name
	parts := strings.Split(name, ",")
	firstName := strings.TrimSpace(parts[0])
	if firstName != "" {
		return firstName, nil
	}
	return name, nil
}
