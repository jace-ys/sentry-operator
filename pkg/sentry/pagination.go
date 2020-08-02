package sentry

import (
	"strconv"
	"strings"
)

type Page struct {
	URL     string
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

		resultsBool := strings.Trim(strings.Split(strings.TrimSpace(segments[2]), "=")[1], `"`)
		results, err := strconv.ParseBool(resultsBool)
		if err != nil {
			results = false
		}

		page := &Page{URL: pageURL, Results: results}
		if strings.TrimSpace(segments[1]) == `rel="next"` {
			r.NextPage = page
		} else {
			r.PrevPage = page
		}
	}
}
