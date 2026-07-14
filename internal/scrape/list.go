package scrape

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const listDateLayout = "2006-01-02"

func videoIDFromPath(href string) (string, error) {
	return refIDFromHref(href, "v")
}

func ParseList(doc *goquery.Document, page int) (Result[ListPage], error) {
	result := Result[ListPage]{Value: ListPage{Items: []Summary{}, Number: page}}

	items := doc.Find(selectorListItem)
	if items.Length() == 0 {
		if doc.Find(selectorEmptyState).Length() > 0 {
			return result, fmt.Errorf("%w: list page has no items", ErrEmpty)
		}
		return result, fmt.Errorf("%w: list page has no recognizable items", ErrParse)
	}

	var parseErr error
	items.EachWithBreak(func(_ int, item *goquery.Selection) bool {
		summary, warnings, err := parseListItem(item)
		if err != nil {
			parseErr = err
			return false
		}
		result.Value.Items = append(result.Value.Items, summary)
		result.Warnings = append(result.Warnings, warnings...)
		return true
	})
	if parseErr != nil {
		return result, parseErr
	}

	result.Value.HasNext = doc.Find(selectorNextPage).Length() > 0
	return result, nil
}

func parseListItem(item *goquery.Selection) (Summary, []Warning, error) {
	var warnings []Warning

	href, ok := item.Find(".box").Attr("href")
	if !ok {
		return Summary{}, nil, fmt.Errorf("%w: list item missing box link", ErrParse)
	}
	id, err := videoIDFromPath(href)
	if err != nil {
		return Summary{}, nil, err
	}

	titleSel := item.Find(".video-title")
	code := strings.TrimSpace(titleSel.Find("strong").Text())
	fullTitle := strings.TrimSpace(titleSel.Text())
	title := strings.TrimSpace(strings.TrimPrefix(fullTitle, code))
	if code == "" || title == "" {
		return Summary{}, nil, fmt.Errorf("%w: list item missing code or title", ErrParse)
	}

	dateText := strings.TrimSpace(item.Find(".meta").Text())
	var publishedAt time.Time
	if dateText == "" || dateText == "N/A" {
		warnings = append(warnings, Warning{Field: "published_at", Message: fmt.Sprintf("invalid date %q", dateText)})
	} else {
		parsed, err := time.Parse(listDateLayout, dateText)
		if err != nil {
			warnings = append(warnings, Warning{Field: "published_at", Message: fmt.Sprintf("invalid date %q", dateText)})
		} else {
			publishedAt = parsed
		}
	}

	scoreText := strings.TrimSpace(item.Find(".score").Text())
	if scoreText == "" {
		return Summary{}, nil, fmt.Errorf("%w: list item missing score", ErrParse)
	}
	score, err := parseScore(scoreText)
	if err != nil {
		return Summary{}, nil, err
	}

	coverURL, _ := item.Find(".cover img").Attr("src")

	tagText := item.Find(".tags").Text()
	hasSubtitleMagnet := strings.Contains(tagText, "含中字磁鏈")
	hasMagnet := hasSubtitleMagnet || strings.Contains(tagText, "含磁鏈")
	isPlayable := item.Find(".cover .tag-can-play").Length() > 0

	return Summary{
		ID:          id,
		Code:        code,
		Title:       title,
		CoverURL:    strings.TrimSpace(coverURL),
		PublishedAt: publishedAt,
		Score:       score,
		Availability: Availability{
			HasMagnet:         hasMagnet,
			HasSubtitleMagnet: hasSubtitleMagnet,
			IsPlayable:        isPlayable,
		},
	}, warnings, nil
}
