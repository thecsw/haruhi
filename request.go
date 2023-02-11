package haruhi

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type request struct {
	url      string
	method   string
	params   url.Values
	client   *http.Client
	timeout  time.Duration
	deadline *time.Time
	ctx      context.Context
	body     io.Reader
	path     string
	headers  http.Header
}

func URL(url string) *request {
	return &request{
		url:    url,
		method: http.MethodGet,
		client: http.DefaultClient,
		ctx:    context.TODO(),
	}
}

func (r *request) Method(method string) *request {
	r.method = method
	return r
}

func (r *request) Get() (string, error) {
	return r.ResponseString()
}

func (r *request) Post() (string, error) {
	r.method = http.MethodPost
	return r.ResponseString()
}

func (r *request) Put() (string, error) {
	r.method = http.MethodPut
	return r.ResponseString()
}

func (r *request) Delete() (string, error) {
	r.method = http.MethodDelete
	return r.ResponseString()
}

func (r *request) Path(path string) *request {
	r.path = path
	return r
}

func (r *request) Params(params url.Values) *request {
	r.params = params
	return r
}

func (r *request) Client(client *http.Client) *request {
	r.client = client
	return r
}

func (r *request) Timeout(timeout time.Duration) *request {
	if r.deadline == nil {
		r.timeout = timeout
	}
	return r
}

func (r *request) Deadline(deadline time.Time) *request {
	if r.timeout == 0 {
		r.deadline = &deadline
	}
	return r
}

func (r *request) Body(body io.Reader) *request {
	r.body = body
	return r
}

func (r *request) Headers(headers http.Header) *request {
	mergeHeaders(r.headers, headers)
	return r
}

func (r *request) BodyBytes(body []byte) *request {
	return r.Body(bytes.NewReader(body))
}

func (r *request) BodyString(body string) *request {
	return r.Body(strings.NewReader(body))
}

func (r *request) BodyJson(body any) *request {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		panic(err)
	}
	return r.Body(buf)
}
func (r *request) BodyXML(body any) *request {
	buf := new(bytes.Buffer)
	if err := xml.NewEncoder(buf).Encode(body); err != nil {
		panic(err)
	}
	return r.Body(buf)
}

func (r *request) BodyFormData(body url.Values) *request {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	defer writer.Close()
	for key, values := range body {
		writer.WriteField(key, strings.Join(values, ""))
	}
	r.headers.Add("Content-Type", writer.FormDataContentType())
	return r.Body(buf)
}

func (r *request) Request() (*http.Request, context.CancelFunc, error) {
	var cancel context.CancelFunc
	if r.timeout > 0 {
		r.ctx, cancel = context.WithTimeout(r.ctx, r.timeout)
	} else if r.deadline != nil {
		r.ctx, cancel = context.WithDeadline(r.ctx, *r.deadline)
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url+r.path, r.body)
	mergeHeaders(req.Header, r.headers)

	q := req.URL.Query()
	mergeParams(q, r.params)
	req.URL.RawQuery = q.Encode()

	return req, cancel, err
}
