package filter

import "github.com/cycloidio/terraforming/tag"

type Filter struct {
	Tags    []tag.Tag
	Include []string
	Exclude []string

	exclude map[string]struct{}
}

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

func (f *Filter) calculateExludeMap() {
	aux := make(map[string]struct{})

	for _, e := range f.Exclude {
		aux[e] = struct{}{}
	}

	f.exclude = aux
}
