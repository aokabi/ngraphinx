package lib

import "regexp"

type Regexps []*regexp.Regexp

func (rs Regexps) FindMatchStringFirst(s string) (retr *regexp.Regexp, find bool) {
	for _, r := range rs {
		if r.MatchString(s) {
			return r, true
		}
	}
	return nil, false
}
