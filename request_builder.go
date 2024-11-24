package http

import (
	"net/http"
)

type RequestBuilder struct {
	client *Client
	method string
	_URL   string
	header http.Header
	body   interface{}
}

func (r *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	r.header.Add(key, value)

	return r
}

func (r *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	r.header.Set(key, value)

	return r
}

func (r *RequestBuilder) Body(body interface{}) *RequestBuilder {
	r.body = body

	return r
}

func (r *RequestBuilder) Build() (req *Request, err error) {
	req, err = NewRequest(r.method, r._URL, r.body)
	if err != nil {
		return
	}

	req.Request.Header = r.header

	return
}

func (r *RequestBuilder) Send() (res *http.Response, err error) {
	req, err := r.Build()
	if err != nil {
		return
	}

	res, err = r.client.Do(req)

	return
}

func NewRequestBuilder(client *Client, method, URL string) (builder *RequestBuilder) {
	builder = &RequestBuilder{}

	builder.client = client
	builder.method = method

	if client.BaseURL != "" {
		URL = client.BaseURL + URL
	}

	builder._URL = URL
	builder.header = make(http.Header)

	for k, v := range client.Headers {
		builder.header.Set(k, v)
	}

	return
}
