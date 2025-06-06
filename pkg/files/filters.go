package files

import "github.com/gobwas/glob"

// Filters represents what files to include, and which to exclude from copying operations.
type Filters struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

func (f Filters) Matcher() Matcher {
	m := Matcher{
		Include: make([]glob.Glob, 0, len(f.Include)),
		Exclude: make([]glob.Glob, 0, len(f.Exclude)),
	}
	separators := []rune{'/'}
	for _, p := range f.Include {
		g := glob.MustCompile(p, separators...)
		m.Include = append(m.Include, g)
	}
	for _, p := range f.Exclude {
		g := glob.MustCompile(p, separators...)
		m.Exclude = append(m.Exclude, g)
	}
	return m
}

type Matcher struct {
	Include []glob.Glob
	Exclude []glob.Glob
}

func (m Matcher) Matches(pth string) bool {
	result := false
	for _, pattern := range m.Include {
		if pattern.Match(pth) {
			result = true
		}
	}
	if !result {
		return false
	}
	for _, pattern := range m.Exclude {
		if pattern.Match(pth) {
			return false
		}
	}
	return true
}
