package wfl

import (
	"github.com/dgruber/drmaa2interface"
	"github.com/mitchellh/copystructure"
)

// Template is a higher level job template for simplifying creating dynamically
// JobTemplates.
type Template struct {
	Jt        drmaa2interface.JobTemplate
	iterators map[string]Iterator
	mappers   map[string]Iterator
}

// Iterator is a function which transforms a JobTemplate when called.
type Iterator func(drmaa2interface.JobTemplate) drmaa2interface.JobTemplate

// NewTemplate creates a Template out of a drmaa2interface.JobTemplate
func NewTemplate(jt drmaa2interface.JobTemplate) *Template {
	return &Template{Jt: jt,
		iterators: make(map[string]Iterator, 16),
		mappers:   make(map[string]Iterator, 16)}
}

// AddIterator registers an interation function which transforms the
// internal JobTemplate into another JobTemplate. The function is called
// each time when Next() is called. Multiple Iterators can be registered.
// The execution order or the Iterators is undefined and does not depend
// on the registration order.
func (t *Template) AddIterator(name string, itr Iterator) *Template {
	t.iterators[name] = itr
	return t
}

// Next applies all registered Iterators to the internal job template
// and returns the next version of the job template.
func (t *Template) Next() drmaa2interface.JobTemplate {
	for _, iter := range t.iterators {
		t.Jt = iter(t.Jt)
	}
	return t.Jt
}

// NextMap applies all registered Iterators to the internal job template
// and finally does a temporary mapping of the job template with the
// mapping function specified.
func (t *Template) NextMap(name string) drmaa2interface.JobTemplate {
	t.Next()
	return t.MapTo(name)
}

// AddMap registers a mapping function (same as Iterator) which converts
// the underlying DRMAA2 JobTemplate into a specific form. In difference
// to the iterator functions it does not make any persistent changes to
// the job template. Its intention is to cover the differencens required
// in the job template so that a job can run on different backends.
func (t *Template) AddMap(name string, f Iterator) *Template {
	t.mappers[name] = f
	return t
}

// MapTo transforms the JobTemplate and returns it. It does not make
// changes to the underlying Template.
func (t *Template) MapTo(system string) drmaa2interface.JobTemplate {
	f, ok := t.mappers[system]
	if ok {
		newTemplate, err := copystructure.Copy(t.Jt)
		if err != nil {
			return t.Jt
		}
		return f(newTemplate.(drmaa2interface.JobTemplate))
	}
	return t.Jt
}
