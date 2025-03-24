package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

// ReusableReadCloser implements the io.ReadCloser interface by wrapping an in‑memory
// byte slice and a bytes.Reader. It allows repeated reads of the same data by automatically
// resetting the reader's position to the beginning when the end of the data is reached.
//
// Fields:
//   - mu (sync.Mutex): A mutex to ensure thread-safe operations on the internal state.
//   - data ([]byte): The complete byte slice containing the data to be read.
//   - reader (*bytes.Reader): A bytes.Reader that facilitates reading from the data.
type ReusableReadCloser struct {
	mu     sync.Mutex
	data   []byte
	reader *bytes.Reader
}

// Read reads up to len(p) bytes from the internal data into the provided buffer p.
// It implements the io.Reader interface. When the end of the data is reached (EOF),
// the reader automatically resets to the beginning, allowing the data to be read repeatedly.
//
// Behavior:
//  1. If the internal data is empty, Read returns 0 and no error.
//  2. Data is read from the bytes.Reader into p.
//  3. When an EOF is encountered:
//     - If some data has been read (n > 0), the reader resets and the EOF error is suppressed.
//     - If no data was read (n == 0), the reader resets and the read is retried.
//
// Concurrency:
//   - The method employs a mutex (mu) to guarantee that concurrent calls to Read are thread-safe.
//
// Parameters:
//   - p ([]byte): The buffer into which data is to be read.
//
// Returns:
//   - n (int): The number of bytes successfully read into the buffer p.
//   - err (error): An error encountered during reading (other than EOF, which is handled internally).
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

// reset repositions the internal bytes.Reader to the beginning of the data.
// This helper function is invoked when EOF is encountered to allow the data
// to be read again from the start. Any errors from Seek are ignored because
// the operation is guaranteed to succeed with an in‑memory data source.
func (r *ReusableReadCloser) reset() {
	_, _ = r.reader.Seek(0, io.SeekStart)
}

// Close implements the io.Closer interface. Since ReusableReadCloser operates solely
// on in-memory data and does not manage external resources, Close is a no-op that always returns nil.
//
// Returns:
//   - err (error): Always nil.
func (r *ReusableReadCloser) Close() (err error) {
	return
}

// NewReusableReadCloser creates a new instance of ReusableReadCloser from a variety of input data types.
// The function converts the provided input into an in‑memory byte slice and initializes a bytes.Reader,
// enabling repeated reads of the same data.
//
// Supported input types via type assertion include:
//   - nil: Results in an empty byte slice.
//   - []byte: Uses the provided byte slice directly.
//   - *[]byte: Dereferences the pointer to obtain the byte slice.
//   - string: Converts the string into a byte slice.
//   - *bytes.Buffer: Retrieves the underlying bytes from the buffer.
//   - *bytes.Reader, *strings.Reader, *io.SectionReader, io.ReadSeeker, io.Reader:
//     Reads the entire content into a byte slice using io.ReadAll.
//
// Parameters:
//   - raw (interface{}): The input data to be converted to a byte slice. The function handles
//     the conversion based on the dynamic type of the input.
//
// Returns:
//   - reusableReadCloser (*ReusableReadCloser): A pointer to the newly created ReusableReadCloser instance.
//   - err (error): An error if the input type is unsupported or if an error occurs during data reading.
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

// errUnsupportedType is a package-level error indicating that the provided input type
// to NewReusableReadCloser is not supported.
var errUnsupportedType = errors.New("unsupported type")
