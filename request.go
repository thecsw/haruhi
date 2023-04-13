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

// Request is an internal "staging" struct used when we want
// to track of what kind of Request user wants to make.
type Request struct {
	// Context to use in the request, defaults to `context.TODO()`.
	ctx context.Context
	// Body to pass in the request, defaults to `nil` (no body).
	body io.Reader
	// ifNotStatusCodeHandler is called if the status code is not the one
	// we expect.
	ifNotStatusCodeHandler func(*http.Response) error
	// HTTP client to use, defaults to `http.DefaultClient`.
	client *http.Client
	// Parameters to use in the request's search parameters.
	params url.Values
	// Headers to use in the request (will overwrite the defaults).
	headers http.Header
	// Deadline for the request, defaults to `nil` (no deadline).
	deadline *time.Time
	// errorHandler is called if set by the user.
	errorHandler func(*http.Response, error) error
	// statusCodeHandlers gets invoked on given status codes.
	statusCodeHandlers map[int]func(*http.Response) error
	// URL that user wants to query, should include schema + domain.
	url string
	// Path to look up on the given URL.
	path string
	// Method to use for HTTP request, defaults to "GET".
	method string
	// Username for basic auth.
	username string
	// Password for basic auth.
	password string
	// statusExpectation is the status code we expect to get.
	statusExpectation int
	// Timeout for the request, defaults to 0 (meaning no timeout).
	timeout time.Duration
}

// URL will start building a request with the given URL (scheme+domain),
// an example is `https://go.dev` (notice without the path or parameters).
func URL(what string) *Request {
	return &Request{
		url:                what,
		method:             http.MethodGet,
		client:             http.DefaultClient,
		ctx:                context.TODO(),
		headers:            http.Header{},
		params:             url.Values{},
		statusCodeHandlers: map[int]func(*http.Response) error{},
	}
}

// Method to use in HTTP request, defaults to "GET" -- recommend using
// `http.Method...` constants.
func (r *Request) Method(method string) *Request {
	r.method = method
	return r
}

// Path to navigate to, leading slash will be added if missing.
func (r *Request) Path(path string) *Request {
	r.path = path
	return r
}

// Context will overwrite the current context with given value.
func (r *Request) Context(ctx context.Context) *Request {
	if shouldSetOrPanic(ctx, "context") {
		r.ctx = ctx
	}
	return r
}

// Parameters to use in the URL.
func (r *Request) Params(params url.Values) *Request {
	mergeParams(r.params, params)
	return r
}

// Parameters to use in the URL.
func (r *Request) Param(name, value string) *Request {
	r.params.Add(name, value)
	return r
}

// HTTP client to use, defaults to `http.DefaultClient`.
func (r *Request) Client(client *http.Client) *Request {
	if shouldSetOrPanic(client, "client") {
		r.client = client
	}
	return r
}

// Timeout for the request, defaults to 0 (meaning no timeout).
func (r *Request) Timeout(timeout time.Duration) *Request {
	if shouldSetOrPanic(timeout, "timeout") {
		r.timeout = timeout
	}
	return r
}

// Deadline for the request (absolute time), defaults to `nil`
// (no deadline).
func (r *Request) Deadline(deadline *time.Time) *Request {
	if r.timeout == 0 {
		r.deadline = deadline
	}
	return r
}

// Headers will record headers to use in the request, will override defaults.
func (r *Request) Headers(headers http.Header) *Request {
	mergeHeaders(r.headers, headers, false)
	return r
}

// Headers will record headers to use in the request, will override defaults.
func (r *Request) Header(name, value string) *Request {
	r.headers.Add(name, value)
	return r
}

// Body tells us we need to read the body request from the reader.
func (r *Request) Body(body io.Reader) *Request {
	r.body = body
	return r
}

// BodyBytes will use slice of bytes as body.
func (r *Request) BodyBytes(body []byte) *Request {
	return r.Body(bytes.NewReader(body))
}

// BodyString will use string as body.
func (r *Request) BodyString(body string) *Request {
	return r.Body(strings.NewReader(body))
}

// ErrorHandler will set the error handler to be called if the request
// fails.
func (r *Request) ErrorHandler(errorHandler func(*http.Response, error) error) *Request {
	if shouldSetOrPanic(errorHandler, "handler") {
		r.errorHandler = errorHandler
	}
	r.errorHandler = errorHandler
	return r
}

// StatusCodeHandler will set the handler to be called if the request
// returns given status code.
func (r *Request) StatusCodeHandler(statusCode int, handler func(*http.Response) error) *Request {
	if shouldSetOrPanic(handler, "handler") {
		r.statusCodeHandlers[statusCode] = handler
	}
	return r
}

// IfNotExpectedStatusCode will set the handler to be called if the request
// does not return the expected status code.
func (r *Request) IfNotExpectedStatusCode(statusCode int, handler func(*http.Response) error) *Request {
	if shouldSetOrPanic(handler, "handler") {
		r.statusExpectation = statusCode
		r.ifNotStatusCodeHandler = handler
	}
	return r
}

// BasicAuth sets the request's Authorization header to use HTTP
// Basic Authentication with the provided username and password.
func (r *Request) BasicAuth(username, password string) *Request {
	r.username = username
	r.password = password
	return r
}

// BodyXML will encode given interfact/instance into JSON and use that as body.
func (r *Request) BodyJson(body any) *Request {
	if body == nil {
		return r
	}
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
func (r *Request) BodyXML(body any) *Request {
	if body == nil {
		return r
	}
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
func (r *Request) BodyFormData(body url.Values) *Request {
	if body == nil {
		return r
	}
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
func (r *Request) Request() (*http.Request, context.CancelFunc, error) {
	var cancel context.CancelFunc
	r.ctx, cancel = context.WithCancel(r.ctx)
	if r.timeout > 0 {
		r.ctx, cancel = context.WithTimeout(r.ctx, r.timeout)
	} else if r.deadline != nil {
		r.ctx, cancel = context.WithDeadline(r.ctx, *r.deadline)
	}

	req, err := http.NewRequestWithContext(r.ctx, r.method, r.url+r.path, r.body)
	if req == nil || err != nil {
		return req, cancel, errors.Wrap(err, "haruhi failed to create request")
	}
	mergeHeaders(req.Header, r.headers, true)

	if len(r.username) > 0 || len(r.password) > 0 {
		req.SetBasicAuth(r.username, r.password)
	}

	q := req.URL.Query()
	mergeParams(q, r.params)
	req.URL.RawQuery = q.Encode()

	return req, cancel, err
}

// shouldPanic is a helper function to check if we should panic or not.
func shouldSetOrPanic(src any, name string) bool {
	if src == nil {
		if shouldPanic {
			panic(name + " cannot be nil")
		}
		return false
	}
	return true
}
