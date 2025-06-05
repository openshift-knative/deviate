package files

import "path"

// Filters represents what files to include, and which to exclude from copying operations.
type Filters struct {
	Include []string `json:"include"`
	Exclude []string `json:"exclude"`
}

func (f Filters) Matches(pth string) bool {
	result := false
	for _, pattern := range f.Include {
		matched, _ := path.Match(pattern, pth)
		if matched {
			result = true
		}
	}
	if !result {
		return false
	}
	for _, pattern := range f.Exclude {
		matched, _ := path.Match(pattern, pth)
		if matched {
			return false
		}
	}
	return true
}
