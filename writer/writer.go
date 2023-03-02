package writer

import "github.com/cycloidio/terracognita/interpolator"

//go:generate mockgen -destination=../mock/writer.go -mock_names=Writer=Writer -package mock github.com/cycloidio/terracognita/writer Writer

const (
	// ResourceCategoryKey is an internal key used to specify the category of
	// a resource when writing, it'll be used to select in which file
	// will be written
	ResourceCategoryKey = "tc_category"

	// ModuleCategoryKey is the category used to identify
	// the Module
	ModuleCategoryKey = "tc_module"
)

// Writer it's an interface used to abstract the logic
// of writing results to a Key Value without having to
// deal with types or internal structures
type Writer interface {
	// Write sets the value with the key to the internal structure,
	// the value will be casted to the correct type of each
	// implementation and an error can be returned normally for
	// repeated keys
	Write(key string, value interface{}) error

	// Has checks if the key it's already written
	// or not
	Has(key string) (bool, error)

	// Sync writes the content of the writer
	// to the internal system. Each Writter may have
	// a different implementation of it with different
	// output formats
	Sync() error

	// Interpolate replaces the hardcoded resources link
	// with TF interpolation
	Interpolate(*interpolator.Interpolator)
}
