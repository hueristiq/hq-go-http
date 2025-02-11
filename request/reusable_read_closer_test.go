package request_test

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.source.hueristiq.com/http/request"
)

func TestNewReusableReadCloser(t *testing.T) {
	t.Parallel()

	const inputString = "hello world"

	inputByteSlice := []byte(inputString)

	testCases := []struct {
		name        string
		input       interface{}
		expectedOut string
		expectedErr bool
	}{
		{"Nil Input", nil, "", false},
		{"Byte Slice", inputByteSlice, inputString, false},
		{"Pointer to Byte Slice", &inputByteSlice, inputString, false},
		{"String", inputString, inputString, false},
		{"*bytes.Buffer", bytes.NewBufferString(inputString), inputString, false},
		{"*bytes.Reader", bytes.NewReader(inputByteSlice), inputString, false},
		{"*strings.Reader", strings.NewReader(inputString), inputString, false},
		{"*io.SectionReader", io.NewSectionReader(bytes.NewReader(inputByteSlice), 0, int64(len(inputString))), inputString, false},
		{"io.ReadSeeker", bytes.NewReader(inputByteSlice), inputString, false},
		{"io.Reader", strings.NewReader(inputString), inputString, false},
		{"Invalid Input", 123, "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader, err := request.NewReusableReadCloser(tc.input)

			if tc.expectedErr {
				require.Error(t, err, "Expected an error but got none")

				return
			}

			require.NoError(t, err, "Unexpected error occurred")

			buf := make([]byte, len(tc.expectedOut))
			n, err := reader.Read(buf)

			require.NoError(t, err, "Unexpected read error")

			assert.Equal(t, tc.expectedOut, string(buf[:n]), "Output does not match expected value")
		})
	}
}

func TestReusableReadCloser_ReadRepetition(t *testing.T) {
	t.Parallel()

	input := "hello world"

	rrc, err := request.NewReusableReadCloser(input)

	require.NoError(t, err)
	require.NotNil(t, rrc)

	buf := make([]byte, len(input))

	n, err := rrc.Read(buf)

	require.NoError(t, err)

	assert.Equal(t, len(input), n)
	assert.Equal(t, input, string(buf[:n]))

	buf2 := make([]byte, len(input))

	n, err = rrc.Read(buf2)

	require.NoError(t, err)

	assert.Equal(t, len(input), n)
	assert.Equal(t, input, string(buf2[:n]))
}

func TestReusableReadCloser_EmptyData(t *testing.T) {
	t.Parallel()

	rrc, err := request.NewReusableReadCloser([]byte{})

	require.NoError(t, err)
	require.NotNil(t, rrc)

	buf := make([]byte, 10)

	n, err := rrc.Read(buf)

	require.NoError(t, err)

	assert.Equal(t, 0, n)
}

func TestReusableReadCloser_PartialRead(t *testing.T) {
	t.Parallel()

	input := "hello world"

	rrc, err := request.NewReusableReadCloser(input)

	require.NoError(t, err)
	require.NotNil(t, rrc)

	buf1 := make([]byte, 5)

	n, err := rrc.Read(buf1)

	require.NoError(t, err)

	assert.Equal(t, 5, n)
	assert.Equal(t, input[:5], string(buf1[:n]))

	buf2 := make([]byte, 5)

	n, err = rrc.Read(buf2)

	require.NoError(t, err)

	assert.Equal(t, 5, n)
	assert.Equal(t, input[5:10], string(buf2[:n]))

	buf3 := make([]byte, 5)

	n, err = rrc.Read(buf3)

	require.NoError(t, err)

	assert.Equal(t, 1, n)
	assert.Equal(t, input[10:], string(buf3[:n]))

	buf4 := make([]byte, 5)

	n, err = rrc.Read(buf4)

	require.NoError(t, err)

	assert.Equal(t, 5, n)
	assert.Equal(t, input[:5], string(buf4[:n]))
}

func TestReusableReadCloser_ConcurrentRead(t *testing.T) {
	t.Parallel()

	input := "concurrent"

	rrc, err := request.NewReusableReadCloser(input)

	require.NoError(t, err)
	require.NotNil(t, rrc)

	var wg sync.WaitGroup

	numGoroutines := 10

	for range numGoroutines {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for range 5 {
				buf := make([]byte, len(input))

				n, err := rrc.Read(buf)
				if err != nil {
					t.Errorf("unexpected error: %v", err)

					return
				}

				if n != len(input) {
					t.Errorf("expected %d bytes, got %d", len(input), n)

					return
				}

				if string(buf) != input {
					t.Errorf("expected %q, got %q", input, string(buf))

					return
				}
			}
		}()
	}

	wg.Wait()
}
