package writer

// Options given to the writers
type Options struct {
	// Interpolate means the ability to interpolate
	// variables in HCL files or building dependencies
	// in a TFState
	Interpolate bool

	// Module tells the Writers (basically HCL) that will
	// need also to write a Module, and the value is the
	// name it has
	Module string

	// ModuleVariables will be all the keys that we want
	// to use as variables when writing. If empty
	// means use all attributes as variables
	ModuleVariables map[string]struct{}
}

// HasModule will check if the Module is empty or not
func (o Options) HasModule() bool {
	return o.Module != ""
}
