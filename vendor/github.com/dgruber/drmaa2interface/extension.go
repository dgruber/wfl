package drmaa2interface

// Extension is a struct which is embedded in DRMAA2 objects
// which are extensible. The extension is named in the DRMAA2
// spec as SetIntanceValue / GetInstanceValue.
type Extension struct {
	ExtensionList map[string]string // stores the extension requests as string
}

// Extensible is an interface which defines functions used to
// interact with extensible data structures (JobTemplate, JobInfo etc.).
type Extensible interface {
	// ListExtensions lists all implementation specific key names for
	// a particular DRMAA2 extensible data type
	ListExtensions() []string
	DescribeExtension(string) string
}
