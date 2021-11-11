package rustack

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type RustackApiError struct {
	msg  string
	code int
}

func NewRustackApiError(url string, resp *http.Response) error {
	body, _ := ioutil.ReadAll(resp.Body)
	msg := fmt.Sprintf("HTTP request failure on %s:\n%d: %s", url, resp.StatusCode, string(body))

	return &RustackApiError{
		msg:  msg,
		code: resp.StatusCode,
	}
}

func (e *RustackApiError) Error() string   { return e.msg }
func (e *RustackApiError) Message() string { return e.msg }
func (e *RustackApiError) Code() int       { return e.code }
