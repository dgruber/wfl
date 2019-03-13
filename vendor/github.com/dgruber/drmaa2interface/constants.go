// Package drmaa2interface implements the DRMAA2 Go interface.
// Actual Go DRMAA2 compatible implementations should use that
// interface to guarantee compatibility.
package drmaa2interface

import (
	"time"
)

// ZeroTime is a special timeout value: Don't wait
const ZeroTime = time.Duration(0)

// InfiniteTime is a special timeout value: Wait probably infinitly
const InfiniteTime = time.Duration(1<<63 - 1)

// Now is always interpreted as the current time.
// const Now = -2

// UnsetTime is a special time value: Time or date not set
const UnsetTime = -3

// UnsetNum defines an unset number (which is different than 0)
const UnsetNum = -1

// UnsetEnum defines an unset enum (required?)
const UnsetEnum = -1

// UnsetList is used to differentiate between an empty and
// an unspecified (unset) list. Probably not needed in Go.
var UnsetList []interface{}

// UnsetDict is means no dictionary is set (nil)
var UnsetDict map[string]string

// UnsetJobInfo means no job info is set (nil)
var UnsetJobInfo *JobInfo

// PlaceholderHomeDir is a specified string which can be used to
// request the home directory of the user (e.g. as the output
// directory of the job).
const PlaceholderHomeDir = "$DRMAA2_HOME_DIR$"

// PlaceholderWorkingDir is a specified string which can be
// used to point to the current working directory of the job.
const PlaceholderWorkingDir = "$DRMAA2_WORKING_DIR$"

// PlaceholderIndex is a specified string which can be used
// to put the output of a job in a file which has the array
// job task ID within the directory structure or filename.
const PlaceholderIndex = "$DRMAA2_INDEX$"
