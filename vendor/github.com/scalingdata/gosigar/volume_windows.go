package sigar

// #include <stdlib.h>
// #include <windows.h>
import "C"

import (
	"fmt"
	"syscall"
)

type WindowsVolumeIterator struct {
	handle     C.HANDLE
	buffer     []uint16
	bufferSize C.DWORD
	volume     FileSystem
	err        error
	hasNext    bool
}

// Get the first mount point for the volume. If we get an ERROR_MORE_DATA,
// don't return an error, but return the expected buffer size.
// You can attempt to call the method again with the larger buffer size
func getVolumePathName(wideName []uint16, bufferSize uint16) (string, error, uint16) {
	charsLen := C.DWORD(bufferSize)
	expectedBuffer := C.DWORD(0)
	chars := make([]uint16, charsLen)

	if C.GetVolumePathNamesForVolumeNameW((*C.WCHAR)(&wideName[0]), (*C.WCHAR)(&chars[0]), charsLen, &expectedBuffer) == C.FALSE {
		errno := syscall.GetLastError()
		err, ok := errno.(syscall.Errno)
		if !ok || err != C.ERROR_MORE_DATA {
			return "", errno, 0
		} else if err == C.ERROR_MORE_DATA {
			return "", nil, uint16(expectedBuffer)
		}
	}
	volumePaths := syscall.UTF16ToString(chars)
	return volumePaths, nil, 0
}

func NewWindowsVolumeIterator() (*WindowsVolumeIterator, error) {
	bufSize := C.DWORD(C.MAX_PATH)
	buf := make([]uint16, bufSize)

	// Get a handle to iterate over all the attached disks
	diskHandle := C.FindFirstVolumeW((*C.WCHAR)(&buf[0]), bufSize)
	if diskHandle == C.HANDLE(uintptr(0xFF)) {
		err := syscall.GetLastError()
		return nil, fmt.Errorf("Got bad handle when calling FindFirstVolume - %v", err)
	}

	return &WindowsVolumeIterator{
		handle:     diskHandle,
		buffer:     buf,
		bufferSize: bufSize,
		hasNext:    true,
	}, nil
}

// Get the last volume from the iterator
func (self *WindowsVolumeIterator) Volume() FileSystem {
	return self.volume
}

// Get the last error that stopped iteration
func (self *WindowsVolumeIterator) Error() error {
	return self.err
}

/* Iterate over the attached volumes. Returns true if there
   is a volume available from Volume(). Always check Error()
   when you're finished iterating. */
func (self *WindowsVolumeIterator) Next() bool {
	/* Because of the weird structure of `FindFirstVolume`,
	   we have a volume buffered before the first `Next()`. */
	if !self.hasNext {
		return false
	}

	// Convert the currently buffered event to a Go struct
	volume := syscall.UTF16ToString(self.buffer)

	// Get the first mount point for the buffered volume
	volumeMount, err, newSize := getVolumePathName(self.buffer, C.MAX_PATH)
	if newSize > 0 {
		// Retry once if we get an error about buffers being too small
		volumeMount, err, _ = getVolumePathName(self.buffer, newSize)
	}

	if err != nil {
		self.err = fmt.Errorf("Error getting volume mountpoints: %v", err)
		return false
	}

	self.volume = FileSystem{
		DevName: volume,
		DirName: volumeMount,
	}

	// Put the next volume in the buffer
	err, hasNext := self.getNextVolume()
	if !hasNext {
		// If there is no buffered event, the next call to Next() will return false
		self.hasNext = false
	}

	// If there was an error getting the next volume, fail immediately
	if err != nil {
		self.err = err
		return false
	}

	return true
}

func (self *WindowsVolumeIterator) Close() {
	C.FindVolumeClose(self.handle)
}

func (self *WindowsVolumeIterator) getNextVolume() (error, bool) {
	if C.FindNextVolumeW(self.handle, (*C.WCHAR)(&self.buffer[0]), self.bufferSize) == C.FALSE {
		err := syscall.GetLastError()
		errno, _ := err.(syscall.Errno)
		if err == nil || errno == C.ERROR_NO_MORE_FILES {
			// There are no more volumes
			return nil, false
		} else {
			// Unknown error
			return err, false
		}
	}
	return nil, true
}
