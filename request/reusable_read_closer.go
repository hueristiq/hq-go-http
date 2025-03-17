package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ReusableReadCloser implements the io.ReadCloser interface, but with a twist.
// Instead of permanently ending with an EOF when all data is read, it automatically
// resets to the beginning. This makes it possible to repeatedly read the same data,
// which is useful in scenarios like re-sending HTTP request bodies.
//
// It wraps an in-memory byte slice and a bytes.Reader to allow repeatedly reading
// the same data. When the end of the data is reached, the internal reader resets
// to the beginning so that the data can be read again.
//
// Fields:
//   - mu: A mutex to ensure thread-safe operations on the internal state.
//   - data: A byte slice that holds the complete data to be read.
//   - reader: A bytes.Reader that provides the actual reading functionality from the byte slice.
type ReusableReadCloser struct {
	mu     sync.Mutex
	data   []byte
	reader *bytes.Reader
}

// NewReusableReadCloser creates and returns a new instance of ReusableReadCloser from various types of input data.
// It converts the provided input into an in‑memory byte slice and sets up an internal bytes.Reader to support repeated reads.
//
// Supported input types (via type assertion):
//   - nil: Produces an empty byte slice.
//   - []byte: Uses the provided byte slice directly.
//   - *[]byte: Dereferences the pointer to obtain the byte slice.
//   - string: Converts the string into a byte slice.
//   - *bytes.Buffer: Uses the underlying bytes from the buffer.
//   - *bytes.Reader, *strings.Reader, *io.SectionReader, io.ReadSeeker, io.Reader:
//     Reads the entire content into a byte slice using io.ReadAll.
//
// Parameters:
//   - raw (interface{}): The input data to be converted into a byte slice. The dynamic type of raw determines
//     how the conversion is performed.
//
// Returns:
//   - reusableReadCloser (*ReusableReadCloser): A pointer to the newly created ReusableReadCloser instance that wraps the data.
//   - err (error): An error value if the input type is unsupported or if an error occurs while reading data.
func NewReusableReadCloser(raw interface{}) (reusableReadCloser *ReusableReadCloser, err error) {
	var data []byte

	switch v := raw.(type) {
	case nil:
		data = []byte{}
	case []byte:
		data = v
	case *[]byte:
		data = *v
	case string:
		data = []byte(v)
	case *bytes.Buffer:
		data = v.Bytes()
	case *bytes.Reader, *strings.Reader, *io.SectionReader, io.ReadSeeker, io.Reader:
		var r io.Reader

		switch rr := v.(type) {
		case *bytes.Reader:
			r = rr
		case *strings.Reader:
			r = rr
		case *io.SectionReader:
			r = rr
		case io.ReadSeeker:
			r = rr
		case io.Reader:
			r = rr
		default:
			err = fmt.Errorf("%w: %T", errUnsupportedType, v)

			return
		}

		data, err = io.ReadAll(r)
		if err != nil {
			return
		}
	default:
		err = fmt.Errorf("%w: %T", errUnsupportedType, v)

		return
	}

	reusableReadCloser = &ReusableReadCloser{
		data:   data,
		reader: bytes.NewReader(data),
	}

	return
}

// Read reads up to len(p) bytes from the internal data into the provided buffer p.
// It implements the io.Reader interface. If the reader reaches the end of the data (EOF),
// it automatically resets to the beginning so that subsequent read operations continue from the start.
//
// Behavior:
//  1. If the internal data is empty, the method returns zero bytes read and no error.
//  2. It reads data from the internal bytes.Reader into p.
//  3. When an EOF is encountered:
//     - If some data was read before the EOF (n > 0), it resets the reader and suppresses the EOF error.
//     - If no data was read (n == 0), it resets the reader and retries the read operation.
//
// Concurrency:
//   - A mutex (mu) is used to ensure that concurrent invocations of Read are thread-safe.
//
// Parameters:
//   - p ([]byte): The destination buffer into which data will be read.
//
// Returns:
//   - n (int): The number of bytes read into the buffer p.
//   - err (error): An error value if an error (other than EOF, which is handled) occurs during reading.
func (r *ReusableReadCloser) Read(p []byte) (n int, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.data) == 0 {
		return
	}

	n, err = r.reader.Read(p)
	if err == io.EOF {
		if n > 0 {
			r.reset()

			err = nil
		} else {
			r.reset()

			n, err = r.reader.Read(p)
		}
	}

	return
}

// reset repositions the internal bytes.Reader back to the beginning of the data.
// This helper function is called when EOF is encountered to allow repeated reading.
// It uses the Seek method on the bytes.Reader to set the read offset to 0.
// Any errors from Seek are ignored since the operation is guaranteed to succeed
// with an in‑memory data source.
func (r *ReusableReadCloser) reset() {
	_, _ = r.reader.Seek(0, io.SeekStart)
}

// Close implements the io.Closer interface.
// Since ReusableReadCloser only uses in-memory data and does not manage external resources,
// Close is effectively a no-op and always returns nil.
//
// Returns:
//   - err (error): Always nil.
func (r *ReusableReadCloser) Close() (err error) {
	return
}

// errUnsupportedType is a package-level error value used to indicate that the provided
// input type to NewReusableReadCloser is not supported.
var errUnsupportedType = errors.New("unsupported type")
