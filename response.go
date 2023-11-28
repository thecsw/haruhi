package haruhi

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

// Make a blocking GET request and return the response as string.
func (r *Request) Get() (string, error) {
	r.method = http.MethodGet
	return r.ResponseString()
}

// Make a blocking POST request and return the response as string.
func (r *Request) Post() (string, error) {
	r.method = http.MethodPost
	return r.ResponseString()
}

// Make a blocking PUT request and return the response as string.
func (r *Request) Put() (string, error) {
	r.method = http.MethodPut
	return r.ResponseString()
}

// Make a blocking DELETE request and return the response as string.
func (r *Request) Delete() (string, error) {
	r.method = http.MethodDelete
	return r.ResponseString()
}

// Make a non-blocking request and get the response object with cancel.
func (r *Request) Response() (resp *http.Response, cancel context.CancelFunc, err error) {
	var req *http.Request
	req, cancel, err = r.Request()
	if err != nil {
		return
	}
	resp, err = r.client.Do(req)

	// Call the error handler if it has been set.
	if err != nil {
		if r.errorHandler != nil {
			err = r.errorHandler(resp, err)
		}
		return
	}
	// What if the response was nil?
	if resp == nil {
		err = errors.New("nil response from http client")
		return
	}
	// If a handler for this status code is registered, call it.
	if codeHandler, found := r.statusCodeHandlers[resp.StatusCode]; found {
		err = codeHandler(resp)
		return
	}
	// If a status code expectation is set, check it.
	if r.statusExpectation > 0 && resp.StatusCode != r.statusExpectation {
		err = r.ifNotStatusCodeHandler(resp)
		return
	}
	return
}

// Make a request (parked) and get the body reader (needs closing) with cancel.
func (r *Request) ResponseBody() (body io.ReadCloser, cancel context.CancelFunc, err error) {
	resp, cancel, err := r.Response()
	if err != nil {
		return
	}
	body = resp.Body
	return
}

// Make a request and get the response as `*bytes.Buffer`.
func (r *Request) ResponseBuffer() (*bytes.Buffer, error) {
	body, cancel, err := r.ResponseBody()
	defer cancel()
	if err != nil {
		return nil, err
	}
	defer body.Close()
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(body); err != nil {
		return nil, err
	}
	return buf, nil
}

// Make a request and get the response as bytes.
func (r *Request) ResponseBytes() ([]byte, error) {
	buf, err := r.ResponseBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

// Make a request and get the response as a string.
func (r *Request) ResponseString() (string, error) {
	buf, err := r.ResponseBuffer()
	if err != nil {
		return "", err
	}
	return buf.String(), err
}

// Make a request and decode the JSON response into given interface.
func (r *Request) ResponseJson(v any) error {
	body, cancel, err := r.ResponseBody()
	defer cancel()
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

// Make a request and decode the XML response into given interface.
func (r *Request) ResponseXML(v any) error {
	body, cancel, err := r.ResponseBody()
	defer cancel()
	if err != nil {
		return err
	}
	defer body.Close()
	return xml.NewDecoder(body).Decode(v)
}
