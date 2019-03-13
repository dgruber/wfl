package drmaa2interface

// Queue implements all required elements of a queue.
type Queue struct {
	Extensible
	Extension `xml:"-" json:"-"`
	Name      string `xml:"name"`
}
