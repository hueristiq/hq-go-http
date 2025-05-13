package http

type Header struct {
	mode  mode
	key   string
	value string
}

type mode string

const (
	HeaderModeAdd mode = "add"
	HeaderModeSet mode = "set"
)

func NewHeader(key, value string, mode mode) (h Header) {
	h = Header{
		key:   key,
		value: value,
		mode:  mode,
	}

	return
}
