package haruhi

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

func (r *request) Response() (body io.ReadCloser, cancel context.CancelFunc, err error) {
	var req *http.Request
	req, cancel, err = r.Request()
	if err != nil {
		return
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	body = resp.Body
	return
}

func (r *request) ResponseBuffer() (*bytes.Buffer, error) {
	body, cancel, err := r.Response()
	defer cancel()
	if err != nil {
		return nil, err
	}
	defer body.Close()
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, body); err != nil && err != io.EOF {
		return buf, err
	}
	return buf, nil
}

func (r *request) ResponseBytes() ([]byte, error) {
	buf, err := r.ResponseBuffer()
	return buf.Bytes(), err
}

func (r *request) ResponseString() (string, error) {
	buf, err := r.ResponseBuffer()
	return buf.String(), err
}

func (r *request) ResponseJson(v any) error {
	body, cancel, err := r.Response()
	defer cancel()
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

func (r *request) ResponseXML(v any) error {
	body, cancel, err := r.Response()
	defer cancel()
	if err != nil {
		return err
	}
	defer body.Close()
	return xml.NewDecoder(body).Decode(v)
}
