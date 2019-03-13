package drmaa2os

type DRMAA2Error struct {
	message string
}

func (d DRMAA2Error) Error() string {
	return d.message
}

var (
	ErrorUnsupportedOperation = DRMAA2Error{"This optional function is not suppported."}
	ErrorJobNotExists         = DRMAA2Error{"The job does not exist."}
	ErrorInvalidState         = DRMAA2Error{"Invalid state."}
	ErrorInternal             = DRMAA2Error{"Internal error occurred."}
	ErrorInvalidSession       = DRMAA2Error{"The session used for the method call is not valid."}
)
