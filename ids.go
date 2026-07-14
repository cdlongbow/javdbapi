package javdbapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type VideoID string
type ActorID string
type MakerID string
type DirectorID string
type SeriesID string

var resourceIDPattern = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func parseResourceID(kind, raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if !resourceIDPattern.MatchString(raw) {
		return "", fmt.Errorf("%w: invalid %s id %q", ErrInvalidQuery, kind, raw)
	}
	return raw, nil
}

// s2tMap maps common simplified Chinese characters to their traditional counterparts.
var s2tMap = map[rune]rune{
	'斋': '齋', '咏': '詠', '艺': '藝', '冯': '馮', '陈': '陳', '驿': '驛',
	'苍': '蒼', '罗': '羅', '条': '條', '东': '東', '么': '麼', '广': '廣',
	'发': '發', '义': '義', '乌': '烏', '书': '書', '务': '務', '备': '備',
	'复': '復', '与': '與', '丑': '醜', '冠': '冠', '刘': '劉', '刊': '刊',
	'卫': '衛', '印': '印', '危': '危', '却': '卻', '卷': '捲', '刷': '刷',
	'动': '動', '助': '助', '南': '南', '协': '協', '单': '單', '卖': '賣',
	'围': '圍', '图': '圖', '国': '國', '团': '團', '园': '園',
	'钟': '鐘', '银': '銀', '镜': '鏡', '长': '長', '门': '門', '间': '間',
	'阵': '陣', '队': '隊', '阳': '陽', '阴': '陰', '难': '難',
	'响': '響', '领': '領', '页': '頁', '风': '風', '飞': '飛', '馆': '館',
	'龙': '龍', '龟': '龜', '麦': '麥', '麻': '麻', '麽': '麼', '黄': '黃',
	'鱼': '魚', '鸟': '鳥', '鸡': '雞', '马': '馬', '骨': '骨', '高': '高',
	'鬼': '鬼', '齐': '齊', '韩': '韓', '驱': '驅',
	'岛': '島', '爱': '愛', '优': '優',
	'樱': '櫻', '乡': '鄉', '枫': '楓',
	'泽': '澤', '结': '結', '桥': '橋', '圣': '聖',
	'宫': '宮', '绪': '緒', '铃': '鈴',
}

// toTraditional converts a simplified Chinese string to traditional Chinese
// using a character-by-character mapping.
func toTraditional(s string) string {
	runes := []rune(s)
	for i, r := range runes {
		if tr, ok := s2tMap[r]; ok {
			runes[i] = tr
		}
	}
	return string(runes)
}

// aliasMap handles common character substitutions for Japanese name
// transcription. These are not simplified/traditional differences but
// Japanese-specific character variants commonly used in Chinese
// transliteration of Japanese names. Applied after s2tMap so that
// e.g. 樱→櫻→桜 chains correctly.
var aliasMap = map[rune]rune{
	'筱': '篠',
	'穗': '穂',
	'理': '裏',
	'户': '戸',
}

// NormalizeName applies both simplified-to-traditional conversion and
// character alias substitution to produce a name suitable for search.
func NormalizeName(s string) string {
	runes := []rune(s)
	// First pass: simplified → traditional
	for i, r := range runes {
		if tr, ok := s2tMap[r]; ok {
			runes[i] = tr
		}
	}
	// Second pass: character aliases (may target results of first pass)
	for i, r := range runes {
		if alias, ok := aliasMap[r]; ok {
			runes[i] = alias
		}
	}
	return string(runes)
}

func ParseVideoID(raw string) (VideoID, error) {
	value, err := parseResourceID("video", raw)
	if err != nil {
		return "", err
	}
	return VideoID(value), nil
}

// ResolveVideoID first tries to parse the raw string as a video ID. If that
// fails (e.g. the caller supplied a code like "DLDSS-271"), it performs a
// search to look up the internal ID.
func (c *Client) ResolveVideoID(ctx context.Context, raw string) (VideoID, error) {
	id, err := ParseVideoID(raw)
	if err == nil {
		return id, nil
	}
	return c.resolveVideoIDByCode(ctx, raw)
}

func (c *Client) resolveVideoIDByCode(ctx context.Context, code string) (VideoID, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return "", fmt.Errorf("%w: empty code", ErrNotFound)
	}
	page, err := c.Search(ctx, SearchQuery{Keyword: code, Page: 1})
	if err != nil {
		return "", err
	}
	if len(page.Items) == 0 {
		return "", fmt.Errorf("%w: no video found for code %q", ErrNotFound, code)
	}
	return page.Items[0].ID, nil
}

func ParseActorID(raw string) (ActorID, error) {
	value, err := parseResourceID("actor", raw)
	return ActorID(value), err
}

func ParseMakerID(raw string) (MakerID, error) {
	value, err := parseResourceID("maker", raw)
	return MakerID(value), err
}

func ParseDirectorID(raw string) (DirectorID, error) {
	value, err := parseResourceID("director", raw)
	return DirectorID(value), err
}

func ParseSeriesID(raw string) (SeriesID, error) {
	value, err := parseResourceID("series", raw)
	return SeriesID(value), err
}
