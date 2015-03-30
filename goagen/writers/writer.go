package writers

// Code generation writer interface
type Writer interface {
	// FunctionName gives the name of the function that generates the artefact (code,
	// documentation, client etc.) for a given resource.
	// The function signature must be:
	//     func (resource *design.Resource) error
	FunctionName() string
	// Source of function that generates the artefact and supporting code if any.
	// This source is written inline in the generator main go source file.
	Source() string
}
