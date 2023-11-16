package rustack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RustackApiError struct {
	msg          string
	code         int
	body         []byte
	errorAliases []string
}

func NewRustackApiError(url string, resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	msg := fmt.Sprintf("HTTP request failure on %s:\n%d: %s", url, resp.StatusCode, string(body))
	var parsedBody struct {
		ErrorAliases []string `json:"error_alias"`
	}
	json.Unmarshal(body, &parsedBody)
	return &RustackApiError{
		msg:          msg,
		code:         resp.StatusCode,
		body:         body,
		errorAliases: parsedBody.ErrorAliases,
	}
}

func (e *RustackApiError) Error() string          { return e.msg }
func (e *RustackApiError) Message() string        { return e.msg }
func (e *RustackApiError) Code() int              { return e.code }
func (e *RustackApiError) Body() []byte           { return e.body }
func (e *RustackApiError) ErrorAliases() []string { return e.errorAliases }
