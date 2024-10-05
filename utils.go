package http

import (
	"io"

	hqgoreaderutil "github.com/hueristiq/hqgoutils/reader"
)

type ContextOverride string

const (
	RetryMax ContextOverride = "retry-max"
)

func getReusableBodyandContentLength(rawBody interface{}) (reader *hqgoreaderutil.ReusableReadCloser, length int64, err error) {
	if rawBody != nil {
		switch body := rawBody.(type) {
		// If they gave us a function already, great! Use it.
		case hqgoreaderutil.ReusableReadCloser:
			reader = &body
		case *hqgoreaderutil.ReusableReadCloser:
			reader = body
		// If they gave us a reader function read it and get reusablereader
		case func() (io.Reader, error):
			var tmp io.Reader

			tmp, err = body()
			if err != nil {
				return
			}

			reader, err = hqgoreaderutil.NewReusableReadCloser(tmp)
			if err != nil {
				return
			}
		// If ReusableReadCloser is not given try to create new from it
		// if not possible return error
		default:
			reader, err = hqgoreaderutil.NewReusableReadCloser(body)
			if err != nil {
				return
			}
		}
	}

	if reader != nil {
		length, err = getReaderLength(reader)
		if err != nil {
			return
		}
	}

	return
}
