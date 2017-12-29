package sigar

/* These are convenience methods for extracting metrics from the Windows
   Performance Data Helper (PDH) library */

import (
	"fmt"
	"github.com/scalingdata/win"
)

// An error encountered calling PDH
type PdhError struct {
	errno  uint32
	method string
}

func (self *PdhError) Error() string {
	return fmt.Sprintf("Error calling %v (%v): %v", self.method, self.errno, win.PdhFormatError(self.errno))
}

func NewPdhError(method string, errno uint32) *PdhError {
	return &PdhError{
		errno:  errno,
		method: method,
	}
}

// An error encountered working with a specific PDH counter
type PdhCounterError struct {
	errno  uint32
	method string
	query  string
}

func NewPdhCounterError(method, query string, errno uint32) *PdhCounterError {
	return &PdhCounterError{
		errno:  errno,
		method: method,
		query:  query,
	}
}

func (self *PdhCounterError) Error() string {
	return fmt.Sprintf("Error calling %v on counter %v (%v): %v", self.method, self.query, self.errno, win.PdhFormatError(self.errno))
}

/* Run the set of provided queries once and report back the raw counter values for each.
   Queries must have a wildcard, like `\Processor(*)\% idle time`. Returns a map of
   wildcard values to arrays of results. The indices in the array match up with the
   index of the query in `queries`. */
func runRawPdhArrayQueries(queries []string) (map[string][]uint64, error) {
	var query win.PDH_HQUERY
	counters := make([]win.PDH_HCOUNTER, len(queries))
	results := make(map[string][]uint64, len(queries))
	// Open a new query
	err := win.PdhOpenQuery(0, 0, &query)
	if err != 0 {
		return nil, NewPdhError("PdhOpenQuery", err)
	}

	// Add all of the counters. If any fail, give up.
	for i := 0; i < len(queries); i++ {
		err = win.PdhAddEnglishCounter(query, queries[i], 0, &counters[i])
		if err != 0 {
			win.PdhCloseQuery(query)
			return nil, NewPdhCounterError("PdhAddEnglishCounter", queries[i], err)
		}
	}

	// Collect query data only once. For raw counters this will give us a value at a point in time
	err = win.PdhCollectQueryData(query)
	if err != 0 {
		win.PdhCloseQuery(query)
		return nil, NewPdhError("PdhCollectQueryData", err)
	}

	// For each query PDH will hand back an array of key-value pairs
	for i := 0; i < len(queries); i++ {
		bufSize := uint32(0)
		items := uint32(0)

		// Get the size for the raw data buffer
		err = win.PdhGetRawCounterArray(counters[i], &bufSize, &items, nil)
		if err != win.PDH_MORE_DATA {
			win.PdhCloseQuery(query)
			return nil, NewPdhCounterError("PdhGetRawCounterArray", queries[i], err)
		}
		buffer := make([]byte, bufSize)
		err = win.PdhGetRawCounterArray(counters[i], &bufSize, &items, &buffer[0])
		if err != 0 {
			win.PdhCloseQuery(query)
			return nil, NewPdhCounterError("PdhGetRawCounterArray", queries[i], err)
		}

		counters := win.PdhConvertRawCounterArray(items, buffer)
		for _, item := range counters {
			keyString := win.UTF16PtrToString(item.SzName)
			results[keyString] = append(results[keyString], uint64(item.RawValue.FirstValue))
		}
	}
	err = win.PdhCloseQuery(query)
	if err != 0 {
		return nil, NewPdhError("PdhCloseQuery", err)
	}
	return results, nil
}

/* Run the set of provided queries once and report back the raw counter values for each.
   Queries cannot have a wildcard. Returns an array of results, where the indices in the array
   match up with the index of the query in `queries`. */
func runRawPdhQueries(queries []string) ([]uint64, error) {
	var query win.PDH_HQUERY
	counters := make([]win.PDH_HCOUNTER, len(queries))
	results := make([]uint64, 0)

	// Open a new query
	err := win.PdhOpenQuery(0, 0, &query)
	if err != 0 {
		return nil, NewPdhError("PdhOpenQuery", err)
	}

	// Add all the counters, give up if any fail
	for i := 0; i < len(queries); i++ {
		err = win.PdhAddEnglishCounter(query, queries[i], 0, &counters[i])
		if err != 0 {
			win.PdhCloseQuery(query)
			return nil, NewPdhCounterError("PdhAddEnglishCounter", queries[i], err)
		}
	}

	// Collect query data at a single point in time. This is sufficient for raw counters.
	err = win.PdhCollectQueryData(query)
	if err != 0 {
		win.PdhCloseQuery(query)
		return nil, NewPdhError("PdhCollectQueryData", err)
	}

	// Each query should return a single struct with a 64-bit int value
	for i := 0; i < len(queries); i++ {
		counterType := uint32(0)
		var counter win.PDH_RAW_COUNTER
		err = win.PdhGetRawCounterValue(counters[i], &counterType, &counter)
		if err != 0 {
			win.PdhCloseQuery(query)
			return nil, NewPdhCounterError("PdhGetRawCounterValue", queries[i], err)
		}
		results = append(results, uint64(counter.FirstValue))
	}
	err = win.PdhCloseQuery(query)
	if err != 0 {
		return nil, NewPdhError("PdhCloseQuery", err)
	}
	return results, nil
}
