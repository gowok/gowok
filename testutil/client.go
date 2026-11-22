package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type Client struct {
	t       testing.TB
	handler http.Handler
}

func New(t testing.TB, handler http.Handler) *Client {
	t.Helper()

	return &Client{t: t, handler: handler}
}

type Response struct {
	t testing.TB
	*httptest.ResponseRecorder
}

func (c *Client) Do(req *http.Request) *Response {
	c.t.Helper()

	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, req)

	return &Response{t: c.t, ResponseRecorder: recorder}
}

func (c *Client) Get(path string) *Response {
	c.t.Helper()

	return c.send(http.MethodGet, path, nil)
}

func (c *Client) Post(path string, body any) *Response {
	c.t.Helper()

	return c.send(http.MethodPost, path, body)
}

func (c *Client) Put(path string, body any) *Response {
	c.t.Helper()

	return c.send(http.MethodPut, path, body)
}

func (c *Client) Patch(path string, body any) *Response {
	c.t.Helper()

	return c.send(http.MethodPatch, path, body)
}

func (c *Client) Delete(path string) *Response {
	c.t.Helper()

	return c.send(http.MethodDelete, path, nil)
}

func (c *Client) send(method string, path string, body any) *Response {
	c.t.Helper()

	reader, contentType := encodeBody(c.t, body)
	req := httptest.NewRequest(method, path, reader)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return c.Do(req)
}

func encodeBody(t testing.TB, body any) (io.Reader, string) {
	t.Helper()

	switch b := body.(type) {
	case nil:
		return nil, ""
	case []byte:
		return bytes.NewReader(b), ""
	case string:
		return strings.NewReader(b), ""
	case io.Reader:
		return b, ""
	default:
		encoded, err := json.Marshal(b)
		if err != nil {
			t.Fatalf("testutil: encode request body: %v", err)
		}
		return bytes.NewReader(encoded), "application/json"
	}
}

func (r *Response) Status() int {
	return r.Code
}

func (r *Response) Bytes() []byte {
	return r.Body.Bytes()
}

func (r *Response) Text() string {
	return r.Body.String()
}

func (r *Response) JSON(dst any) *Response {
	r.t.Helper()

	if err := json.Unmarshal(r.Body.Bytes(), dst); err != nil {
		r.t.Fatalf("testutil: decode response body: %v\nbody: %s", err, r.Text())
	}

	return r
}

func (r *Response) IsStatusEquals(code int) *Response {
	r.t.Helper()

	if r.Code != code {
		r.t.Errorf("testutil: unexpected status: got %d, want %d\nbody: %s", r.Code, code, r.Text())
	}

	return r
}

func (r *Response) IsStatusNotEquals(code int) *Response {
	r.t.Helper()

	if r.Code == code {
		r.t.Errorf("testutil: unexpected status: got %d, want any other\nbody: %s", r.Code, r.Text())
	}

	return r
}

func (r *Response) IsHeaderEquals(key string, want string) *Response {
	r.t.Helper()

	if got := r.Header().Get(key); got != want {
		r.t.Errorf("testutil: unexpected header %q: got %q, want %q", key, got, want)
	}

	return r
}

func (r *Response) IsHeaderNotEquals(key string, want string) *Response {
	r.t.Helper()

	if got := r.Header().Get(key); got == want {
		r.t.Errorf("testutil: unexpected header %q: got %q, want any other", key, got)
	}

	return r
}

func (r *Response) IsBodyEquals(want string) *Response {
	r.t.Helper()

	if got := r.Text(); got != want {
		r.t.Errorf("testutil: unexpected body: got %q, want %q", got, want)
	}

	return r
}

func (r *Response) IsBodyNotEquals(want string) *Response {
	r.t.Helper()

	if got := r.Text(); got == want {
		r.t.Errorf("testutil: unexpected body: got %q, want any other", got)
	}

	return r
}

func (r *Response) IsJSONEquals(want any) *Response {
	r.t.Helper()

	got, err := decodeJSON(r.Body.Bytes())
	if err != nil {
		r.t.Fatalf("testutil: decode response body: %v\nbody: %s", err, r.Text())
	}

	expected, err := decodeJSONValue(want)
	if err != nil {
		r.t.Fatalf("testutil: encode expected JSON: %v", err)
	}

	if !reflect.DeepEqual(got, expected) {
		r.t.Errorf("testutil: unexpected JSON body: got %v, want %v", got, expected)
	}

	return r
}

func (r *Response) IsJSONNotEquals(want any) *Response {
	r.t.Helper()

	got, err := decodeJSON(r.Body.Bytes())
	if err != nil {
		r.t.Fatalf("testutil: decode response body: %v\nbody: %s", err, r.Text())
	}

	expected, err := decodeJSONValue(want)
	if err != nil {
		r.t.Fatalf("testutil: encode expected JSON: %v", err)
	}

	if reflect.DeepEqual(got, expected) {
		r.t.Errorf("testutil: unexpected JSON body: got %v, want any other", got)
	}

	return r
}

func decodeJSON(data []byte) (any, error) {
	var value any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func decodeJSONValue(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return decodeJSON(encoded)
}
