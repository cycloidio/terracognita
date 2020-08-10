package writer

// Options given to the writers
type Options struct {
	// Interpolate means the ability to interpolate
	// variables in HCL files or building dependencies
	// in a TFState
	Interpolate bool
}
