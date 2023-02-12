package haruhi

import "log"

var (
	// logger to use in haruhi for reporting errors.
	logger = log.Default()

	// shouldPanic tells haruhi to panic if intermediate
	// builders return an error.
	shouldPanic = false
)

func init() {
	logger.SetPrefix("[HARUHI]")
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// PanicOnErrors will let all the intermediate functions know that
// if a non-nil fatal error occurs, haruhi should panic and it's
// user's responsibility to recover, otherwise, fallback to defaults.
func PanicOnErrors() {
	shouldPanic = true
}
