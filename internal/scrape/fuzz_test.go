package scrape

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func FuzzParseScore(f *testing.F) {
	f.Add("4.55分, 由3080人評價")
	f.Add("0.0分, 由2人評價")
	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = parseScore(raw)
	})
}

func FuzzParseMagnetMeta(f *testing.F) {
	f.Add("5.76GB, 2個文件")
	f.Add("1.2TB")
	f.Add("2個文件")
	f.Fuzz(func(t *testing.T, raw string) {
		_, _ = parseMagnetMeta(raw)
	})
}

func FuzzParseList(f *testing.F) {
	seed, err := os.ReadFile(filepath.Join("testdata", "list.html"))
	if err != nil {
		f.Fatal(err)
	}
	f.Add(seed)
	f.Fuzz(func(t *testing.T, raw []byte) {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
		if err != nil {
			return
		}
		_, _ = ParseList(doc, 1)
	})
}

func FuzzParseDetail(f *testing.F) {
	for _, name := range []string{"detail-complete.html", "detail-minimal.html"} {
		seed, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
		if err != nil {
			return
		}
		_, _ = ParseDetail(doc, "P9Jkq9")
	})
}

func FuzzParseReviews(f *testing.F) {
	for _, name := range []string{"reviews.html", "reviews-empty.html"} {
		seed, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
		if err != nil {
			return
		}
		_, _ = ParseReviews(doc)
	})
}
