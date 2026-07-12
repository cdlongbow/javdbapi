package javdbapi

import "fmt"

type HomeType string
type HomeFilter string
type HomeSort string

const (
	HomeTypeAll        HomeType = ""
	HomeTypeCensored   HomeType = "censored"
	HomeTypeUncensored HomeType = "uncensored"
	HomeTypeWestern    HomeType = "western"
)

const (
	HomeFilterAll      HomeFilter = "0"
	HomeFilterDownload HomeFilter = "1"
	HomeFilterCNSub    HomeFilter = "2"
	HomeFilterReview   HomeFilter = "3"
)

const (
	HomeSortPublishDate HomeSort = "1"
	HomeSortMagnetDate  HomeSort = "2"
)

func (t HomeType) valid() bool {
	switch t {
	case HomeTypeAll, HomeTypeCensored, HomeTypeUncensored, HomeTypeWestern:
		return true
	default:
		return false
	}
}

func (f HomeFilter) valid() bool {
	switch f {
	case "", HomeFilterAll, HomeFilterDownload, HomeFilterCNSub, HomeFilterReview:
		return true
	default:
		return false
	}
}

func (s HomeSort) valid() bool {
	switch s {
	case "", HomeSortPublishDate, HomeSortMagnetDate:
		return true
	default:
		return false
	}
}

type HomeQuery struct {
	Type   HomeType
	Filter HomeFilter
	Sort   HomeSort
	Page   int
}

type SearchQuery struct {
	Keyword string
	Page    int
}

type MakerFilter string

const (
	MakerFilterAll      MakerFilter = ""
	MakerFilterPlayable MakerFilter = "playable"
	MakerFilterSingle   MakerFilter = "single"
	MakerFilterDownload MakerFilter = "download"
	MakerFilterCNSub    MakerFilter = "cnsub"
	MakerFilterPreview  MakerFilter = "preview"
)

func (f MakerFilter) valid() bool {
	switch f {
	case MakerFilterAll, MakerFilterPlayable, MakerFilterSingle, MakerFilterDownload, MakerFilterCNSub, MakerFilterPreview:
		return true
	default:
		return false
	}
}

type ActorFilter string

const (
	ActorFilterAll      ActorFilter = ""
	ActorFilterPlayable ActorFilter = "p"
	ActorFilterSingle   ActorFilter = "s"
	ActorFilterDownload ActorFilter = "d"
	ActorFilterCNSub    ActorFilter = "c"
)

func (f ActorFilter) valid() bool {
	switch f {
	case ActorFilterAll, ActorFilterPlayable, ActorFilterSingle, ActorFilterDownload, ActorFilterCNSub:
		return true
	default:
		return false
	}
}

type ListSort string

const (
	ListSortPublished    ListSort = "published"
	ListSortRating       ListSort = "rating"
	ListSortPopularity   ListSort = "popularity"
	ListSortWantCount    ListSort = "want_count"
	ListSortWatchedCount ListSort = "watched_count"
)

// sortTypeValue maps to the real page's sort_type=0..4 query parameter.
func (s ListSort) sortTypeValue() (int, error) {
	switch s {
	case "", ListSortPublished:
		return 0, nil
	case ListSortRating:
		return 1, nil
	case ListSortPopularity:
		return 2, nil
	case ListSortWantCount:
		return 3, nil
	case ListSortWatchedCount:
		return 4, nil
	default:
		return 0, fmt.Errorf("%w: invalid sort %q", ErrInvalidQuery, s)
	}
}

type MakerVideosQuery struct {
	MakerID MakerID
	Filter  MakerFilter
	Sort    ListSort
	Page    int
}

type ActorVideosQuery struct {
	ActorID ActorID
	Filters []ActorFilter
	Sort    ListSort
	Page    int
}

type RankingPeriod string
type RankingType string

const (
	RankingPeriodDaily   RankingPeriod = "daily"
	RankingPeriodWeekly  RankingPeriod = "weekly"
	RankingPeriodMonthly RankingPeriod = "monthly"
)

const (
	RankingTypeCensored   RankingType = "censored"
	RankingTypeUncensored RankingType = "uncensored"
	RankingTypeWestern    RankingType = "western"
)

func (p RankingPeriod) valid() bool {
	switch p {
	case RankingPeriodDaily, RankingPeriodWeekly, RankingPeriodMonthly:
		return true
	default:
		return false
	}
}

func (t RankingType) valid() bool {
	switch t {
	case RankingTypeCensored, RankingTypeUncensored, RankingTypeWestern:
		return true
	default:
		return false
	}
}

type RankingQuery struct {
	Period RankingPeriod
	Type   RankingType
	Page   int
}

// normalizeQueryPage maps Page == 0 to the first page and rejects negatives.
func normalizeQueryPage(page int) (int, error) {
	if page < 0 {
		return 0, fmt.Errorf("%w: negative page %d", ErrInvalidQuery, page)
	}
	if page == 0 {
		return 1, nil
	}
	return page, nil
}
