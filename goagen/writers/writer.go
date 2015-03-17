package writers

// Code generation writer interface
type Writer interface {
	// Write content to given output directory
	Write(outputDir string) (*Report, error)
	// Human friendly title
	Title() string
}

// Generation report
type Report struct {
	// Name of generated files
	Generated []string
	// Warnings if any
	Warnings []string
}
