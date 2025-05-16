package http

type Header struct {
	key       string
	value     string
	operation headerOperation
}

type headerOperation string

const (
	headerOperationAppend  headerOperation = "append"
	headerOperationReplace headerOperation = "replace"
)

func NewAddHeader(key, value string) (h Header) {
	h = Header{
		key:       key,
		value:     value,
		operation: headerOperationAppend,
	}

	return
}

func NewSetHeader(key, value string) (h Header) {
	h = Header{
		key:       key,
		value:     value,
		operation: headerOperationReplace,
	}

	return
}
