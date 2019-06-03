package filter

import "github.com/cycloidio/terracognita/tag"

// Filter is the list of all possible
// filters that can be used to filter
// the results
type Filter struct {
	Tags    []tag.Tag
	Include []string
	Exclude []string

	exclude map[string]struct{}
}

// IsExcluded checks if the v is on the Exclude list
func (f *Filter) IsExcluded(v string) bool {
	if len(f.Exclude) == 0 {
		return false
	}

	if f.exclude == nil {
		f.calculateExludeMap()
	}

	_, ok := f.exclude[v]
	return ok
}

// calculateExludeMap makes a map of the Exclude so
// it's easy to operate over them
func (f *Filter) calculateExludeMap() {
	aux := make(map[string]struct{})

	for _, e := range f.Exclude {
		aux[e] = struct{}{}
	}

	f.exclude = aux
}
