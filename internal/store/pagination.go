package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

var DESC = "desc"

type PaginationFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=100"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=ASC DESC"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since" validate:"omitempty,datetime=2006-01-02"`
	Until  string   `json:"until" validate:"omitempty,datetime=2006-01-02"`
}

func (fq *PaginationFeedQuery) Parse(r *http.Request) (*PaginationFeedQuery, error) {
	qs := r.URL.Query()

	limit := qs.Get("limit")
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err == nil {
			fq.Limit = l
		}
		fq.Limit = l
	}

	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err == nil {
			fq.Offset = o
		}
		fq.Offset = o
	}

	sort := qs.Get("sort")
	if sort != "" {
		fq.Sort = sort
	}

	tags := qs.Get("tags")
	if tags != "" {
		fq.Tags = append(fq.Tags, strings.Split(tags, ",")...)
	}

	search := qs.Get("search")
	if search != "" {
		fq.Search = search
	}
	since := qs.Get("since")
	if since != "" {
		fq.Since = ParseTime(since)
	}
	until := qs.Get("until")
	if until != "" {
		fq.Until = ParseTime(until)
	}
	return fq, nil
}

func ParseTime(timeStr string) string {
	t, err := time.Parse(time.DateTime, timeStr)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
