package writer

// Options given to the writers
type Options struct {
	// Interpolate means the ability to interpolate
	// variables in HCL files or building dependencies
	// in a TFState
	Interpolate bool

	// PreSync is the method to call before the `Sync`
	// of the writer in order to perform surgicals modification
	PreSync func(interface{}) error
}
