package mux

import (
	"net/http"
	"regexp"
	"strings"
)

const (
	rankAny = iota
	rankPath
	rankScheme
)

// Matcher types try to match a request.
type Matcher interface {
	Match(*http.Request) bool
	Rank() int
}

// headerMatcher matches the request against header values.
type headerMatcher map[string]comparison

func newHeaderMatcher(pairs ...string) (headerMatcher, error) {
	headers, err := convertStringsToMapString(isEvenPairs, pairs...)
	if err != nil {
		return nil, err
	}

	return headerMatcher(headers), nil
}

func (m headerMatcher) Match(r *http.Request) bool {
	return matchMap(m, r.Header, true)
}

func (m headerMatcher) Rank() int {
	return rankAny
}

// headerRegexMatcher matches the request against header values.
type headerRegexMatcher map[string]comparison

func newHeaderRegexMatcher(pairs ...string) (headerRegexMatcher, error) {
	headers, err := convertStringsToMapRegex(isEvenPairs, pairs...)
	if err != nil {
		return nil, err
	}

	return headerRegexMatcher(headers), nil
}

func (m headerRegexMatcher) Match(r *http.Request) bool {
	return matchMap(m, r.Header, true)
}

func (m headerRegexMatcher) Rank() int {
	return rankAny
}

// MatcherFunc is the function signature used by custom Matchers.
type MatcherFunc func(*http.Request) bool

// Match returns the match for a given request.
func (m MatcherFunc) Match(r *http.Request) bool {
	return m(r)
}

func (m MatcherFunc) Rank() int {
	return rankAny
}

// schemeMatcher matches the request against URL schemes.
type schemeMatcher map[string]struct{}

func newSchemeMatcher(schemes ...string) schemeMatcher {
	schemeMatcher := schemeMatcher{}

	for _, v := range schemes {
		schemeMatcher[strings.ToLower(v)] = struct{}{}
	}

	return schemeMatcher
}

func (m schemeMatcher) Match(r *http.Request) bool {
	if _, found := m[r.URL.Scheme]; found {
		return true
	}

	return false
}

func (m schemeMatcher) Rank() int {
	return rankScheme
}

//pathMatcher matches the request against a URL path.
type pathMatcher string

func (m pathMatcher) Match(r *http.Request) bool {
	if strings.Compare(string(m), r.URL.Path) == 0 {
		return true
	}

	return false
}

func (m pathMatcher) Rank() int {
	return rankPath
}

//pathWithVarsMatcher matches the request against a URL path.
type pathWithVarsMatcher struct {
	regex *regexp.Regexp
}

func newPathWithVarsMatcher(path string) pathWithVarsMatcher {

Loop:
	for {
		switch {
		case strings.Contains(path, ":number") == true:
			path = strings.Replace(path, ":number", "([0-9]{1,})", -1)
			continue
		case strings.Contains(path, ":string") == true:
			path = strings.Replace(path, ":string", "([a-zA-Z]{1,})", -1)
			continue
		default:

			break Loop
		}
	}

	return pathWithVarsMatcher{
		regex: regexp.MustCompile(`^` + path + `$`),
	}
}

func (m pathWithVarsMatcher) Rank() int {
	return rankPath
}

func (m pathWithVarsMatcher) Match(r *http.Request) bool {

	if m.regex.MatchString(r.URL.Path) {
		return true
	}

	return false
}

//pathWithVarsMatcher matches the request against a URL path.
type pathRegexMatcher struct {
	regex *regexp.Regexp
}

func newPathRegexMatcher(path string) pathRegexMatcher {
	path = strings.Replace(path, "#", "", -1)
	return pathRegexMatcher{
		regex: regexp.MustCompile(`^` + path + `$`),
	}
}

func (m pathRegexMatcher) Match(r *http.Request) bool {

	if m.regex.MatchString(r.URL.Path) {
		return true
	}

	return false
}

func (m pathRegexMatcher) Rank() int {
	return rankPath
}

// implements the sort interface (len, swap, less)
// see sort.Sort (Standard Library)
type Matchers []Matcher

func (m Matchers) Len() int {
	return len(m)
}

func (m Matchers) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m Matchers) Less(i, j int) bool {
	return m[i].Rank() < m[j].Rank()
}
