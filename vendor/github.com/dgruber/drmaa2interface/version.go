package drmaa2interface

import (
	"fmt"
)

// Version is a DRMAA2 version type which holds the major version
// and minor version.
type Version struct {
	Major string
	Minor string
}

func (v Version) String() string {
	return fmt.Sprintf("%s.%s", v.Major, v.Minor)
}
