package scrape

import (
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func ParseReviews(doc *goquery.Document) (Result[[]Review], error) {
	result := Result[[]Review]{Value: []Review{}}

	items := doc.Find(selectorReviewItem)
	if items.Length() == 0 {
		if doc.Find(selectorEmptyState).Length() > 0 {
			return result, nil
		}
		return result, fmt.Errorf("%w: reviews page has no recognizable review items", ErrParse)
	}

	var warnings []Warning
	var parseErr error
	items.EachWithBreak(func(_ int, item *goquery.Selection) bool {
		review, ws, err := parseReviewItem(item)
		if err != nil {
			parseErr = err
			return false
		}
		result.Value = append(result.Value, review)
		warnings = append(warnings, ws...)
		return true
	})
	if parseErr != nil {
		return result, parseErr
	}

	result.Warnings = warnings
	return result, nil
}

func parseReviewItem(item *goquery.Selection) (Review, []Warning, error) {
	var warnings []Warning

	idAttr, ok := item.Attr("id")
	if !ok {
		return Review{}, nil, fmt.Errorf("%w: review item missing id", ErrParse)
	}
	id := strings.TrimPrefix(idAttr, "review-item-")
	if id == "" || id == idAttr {
		return Review{}, nil, fmt.Errorf("%w: unexpected review item id %q", ErrParse, idAttr)
	}

	dateText := strings.TrimSpace(item.Find(".time").First().Text())
	if dateText == "" {
		return Review{}, nil, fmt.Errorf("%w: review item missing date", ErrParse)
	}
	publishedAt, err := time.Parse(listDateLayout, dateText)
	if err != nil {
		return Review{}, nil, fmt.Errorf("%w: invalid review date %q: %w", ErrParse, dateText, err)
	}

	// Author is whatever free-standing text remains once every known
	// control/metadata element is stripped from a clone of the item.
	authorSel := item.Clone()
	authorSel.Find(".report.is-pulled-right, .likes.is-pulled-right, .score-stars, .time, .likes-count, .content").Remove()
	author := strings.TrimSpace(authorSel.Text())
	if author == "" {
		return Review{}, nil, fmt.Errorf("%w: review item missing author", ErrParse)
	}

	content := strings.TrimSpace(item.Find(".content").Text())
	if content == "" {
		return Review{}, nil, fmt.Errorf("%w: review item missing content", ErrParse)
	}

	var score *int
	if stars := item.Find(".score-stars i.icon-star"); stars.Length() > 0 {
		n := stars.Not(".gray").Length()
		score = &n
	}

	likes := 0
	if likesText := strings.TrimSpace(item.Find(".likes-count").Text()); likesText != "" {
		n, err := parseCount(likesText)
		if err != nil {
			warnings = append(warnings, Warning{Field: "likes", Message: err.Error()})
		} else {
			likes = n
		}
	}

	return Review{
		ID:          id,
		Author:      author,
		Score:       score,
		PublishedAt: &publishedAt,
		Likes:       likes,
		Content:     content,
	}, warnings, nil
}
