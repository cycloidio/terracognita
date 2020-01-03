package filter

import (
	"fmt"
	"strings"

	"github.com/cycloidio/terracognita/errcode"
	"github.com/cycloidio/terracognita/tag"
	"github.com/pkg/errors"
)

// Filter is the list of all possible
// filters that can be used to filter
// the results
type Filter struct {
	Tags    []tag.Tag
	Include []string
	Exclude []string
	Targets []string

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

// Validate validates that the data inside of the filters is right
func (f *Filter) Validate() error {
	// Validate that the Targets have the right format
	if len(f.Targets) != 0 {
		for _, t := range f.Targets {
			// IDs can have . in between so we validate that at least we have the minimum
			if !strings.Contains(t, ".") {
				return errors.Wrapf(errcode.ErrFilterTargetsInvalid, "the Target %q has an invalid format. The expected format is 'aws_instance.ID'", t)
			}
		}
	}

	return nil
}

// TargetsTypesWithIDs returns all the types (ex: aws_instance) from
// the list of Targets and the IDs
func (f *Filter) TargetsTypesWithIDs() map[string][]string {
	types := make(map[string]map[string]struct{})
	res := make(map[string][]string)

	for _, t := range f.Targets {
		// IDs can have . in between so we validate that at least we have the minimum
		split := strings.SplitN(t, ".", 2)
		ty := split[0]
		id := split[1]

		if _, ok := types[ty]; !ok {
			types[ty] = make(map[string]struct{})
			res[ty] = make([]string, 0)
		}

		if _, ok := types[ty][id]; !ok {
			types[ty][id] = struct{}{}
			res[ty] = append(res[ty], id)
		}
	}

	return res
}

// String returns a stringification of the Filter
func (f *Filter) String() string {
	return fmt.Sprintf(`
	Tags:    %s,
	Include: %s,
	Exclude: %s,
	Targets: %s,
`, f.Tags, f.Include, f.Exclude, f.Targets)
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
