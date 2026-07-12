package scrape

const (
	selectorListItem    = "div.item"
	selectorNextPage    = "a.pagination-next[rel=next]"
	selectorDetailPanel = ".movie-panel-info .panel-block"
	selectorReviewItem  = ".review-item[id^='review-item-']"
	selectorEmptyState  = ".empty-message"
	selectorMagnetItem  = "#magnets-content > .item.columns.is-desktop"
	selectorPlayButton  = ".video-detail .play-button"
)

const (
	labelCode     = "番號:"
	labelDate     = "日期:"
	labelDuration = "時長:"
	labelDirector = "導演:"
	labelMaker    = "片商:"
	labelSeries   = "系列:"
	labelScore    = "評分:"
	labelTags     = "類別:"
	labelActors   = "演員:"
)
