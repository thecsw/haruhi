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

	"github.com/pkg/errors"
)

// request is an internal "staging" struct used when we want
// to track of what kind of request user wants to make.
type request struct {
	// Method to use for HTTP request, defaults to "GET".
	method string

	// URL that user wants to query, should include schema + domain.
	url string

	// Path to look up on the given URL.
	path string

	// Parameters to use in the request's search parameters.
	params url.Values

	// Body to pass in the request, defaults to `nil` (no body).
	body io.Reader

	// Headers to use in the request (will overwrite the defaults).
	headers http.Header

	// HTTP client to use, defaults to `http.DefaultClient`.
	client *http.Client

	// Context to use in the request, defaults to `context.TODO()`.
	ctx context.Context

	// Timeout for the request, defaults to 0 (meaning no timeout).
	timeout time.Duration

	// Deadline for the request, defaults to `nil` (no deadline).
	deadline *time.Time
}

// URL will start building a request with the given URL (scheme+domain),
// an example is `https://go.dev` (notice without the path or parameters).
func URL(url string) *request {
	return &request{
		url:    strings.TrimSuffix(url, "/"),
		method: http.MethodGet,
		client: http.DefaultClient,
		ctx:    context.TODO(),
	}
}

// Method to use in HTTP request, defaults to "GET" -- recommend using
// `http.Method...` constants.
func (r *request) Method(method string) *request {
	r.method = method
	return r
}

// Path to navigate to, leading slash will be added if missing.
func (r *request) Path(path string) *request {
	r.path = addForwardSlash(path)
	return r
}

// Parameters to use in the URL.
func (r *request) Params(params url.Values) *request {
	r.params = params
	return r
}

// HTTP client to use, defaults to `http.DefaultClient`.
func (r *request) Client(client *http.Client) *request {
	r.client = client
	return r
}

// Timeout for the request, defaults to 0 (meaning no timeout).
func (r *request) Timeout(timeout time.Duration) *request {
	if r.deadline == nil {
		r.timeout = timeout
	}
	return r
}

// Deadline for the request (absolute time), defaults to `nil`
// (no deadline).
func (r *request) Deadline(deadline time.Time) *request {
	if r.timeout == 0 {
		r.deadline = &deadline
	}
	return r
}

// Headers will record headers to use in the request, will override defaults.
func (r *request) Headers(headers http.Header) *request {
	mergeHeaders(r.headers, headers, false)
	return r
}

// Body tells us we need to read the body request from the reader.
func (r *request) Body(body io.Reader) *request {
	r.body = body
	return r
}

// BodyBytes will use slice of bytes as body.
func (r *request) BodyBytes(body []byte) *request {
	return r.Body(bytes.NewReader(body))
}

// BodyString will use string as body.
func (r *request) BodyString(body string) *request {
	return r.Body(strings.NewReader(body))
}

// BodyXML will encode given interfact/instance into JSON and use that as body.
func (r *request) BodyJson(body any) *request {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		err = errors.Wrap(err, "couldn't encode into json")
		if shouldPanic {
			logger.Panicln("panic:", err.Error())
		}
		// Otherwise, set the body to nil and log it.
		logger.Println("leaving body empty:", err.Error())
		return r
	}
	return r.Body(buf)
}

// BodyXML will encode given interfact/instance into XML and use that as body.
func (r *request) BodyXML(body any) *request {
	buf := new(bytes.Buffer)
	if err := xml.NewEncoder(buf).Encode(body); err != nil {
		err = errors.Wrap(err, "couldn't encode into xml")
		if shouldPanic {
			logger.Panicln("panic:", err.Error())
		}
		// Otherwise, set the body to nil and log it.
		logger.Println("leaving body empty:", err.Error())
		return r
	}
	return r.Body(buf)
}

// BodyFormData will take values and send them as formdata.
func (r *request) BodyFormData(body url.Values) *request {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	defer writer.Close()
	for key, values := range body {
		if err := writer.WriteField(key, strings.Join(values, "")); err != nil {
			logger.Printf("couldn't write form field %s: %s\n", key, err)
		}
	}
	r.headers.Add("Content-Type", writer.FormDataContentType())
	return r.Body(buf)
}

// Request will build the final haruhi request.
func (r *request) Request() (*http.Request, context.CancelFunc, error) {
	var cancel context.CancelFunc = func() {}
	if r.timeout > 0 {
		r.ctx, cancel = context.WithTimeout(r.ctx, r.timeout)
	} else if r.deadline != nil {
		r.ctx, cancel = context.WithDeadline(r.ctx, *r.deadline)
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url+r.path, r.body)
	mergeHeaders(req.Header, r.headers, true)

	q := req.URL.Query()
	mergeParams(q, r.params)
	req.URL.RawQuery = q.Encode()

	return req, cancel, err
}
