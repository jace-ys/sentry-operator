package sentry

import (
	"strconv"
	"strings"
)

type Page struct {
	URL     string
	Cursor  string
	Results bool
}

func (r *Response) parsePaginationLinks() {
	links := r.Header.Get("Link")
	for _, link := range strings.SplitN(links, ",", 2) {
		segments := strings.SplitN(strings.TrimSpace(link), ";", 4)

		if len(segments) < 3 {
			continue
		}

		pageURL := strings.TrimLeft(strings.TrimSpace(segments[0]), "<")
		pageURL = strings.TrimRight(pageURL, ">")

		resultsBool := parseSegmentValue(segments[2])
		results, err := strconv.ParseBool(resultsBool)
		if err != nil {
			results = false
		}

		page := &Page{
			URL:     pageURL,
			Cursor:  parseSegmentValue(segments[3]),
			Results: results,
		}

		if parseSegmentValue(segments[1]) == "next" {
			r.NextPage = page
		} else {
			r.PrevPage = page
		}
	}
}

func parseSegmentValue(segment string) string {
	parts := strings.Split(strings.TrimSpace(segment), "=")
	if len(parts) < 2 {
		return ""
	}

	return strings.Trim(parts[1], `"`)
}
