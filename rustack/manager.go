package rustack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const DefaultBaseURL = "https://cp.sbcloud.ru"
const RetryTime = 500    // ms
const LockTimeout = 1200 // seconds
const TaskTimeout = 600  // seconds

type Manager struct {
	Client    *http.Client
	ClientID  string
	Logger    logger
	BaseURL   string
	Token     string
	UserAgent string
	ctx       context.Context
}

type ObjectLocked struct {
	Details        []interface{} `json:"details"`
	ErrorAlias     []interface{} `json:"error_alias"`
	NonFieldErrors []interface{} `json:"non_field_errors"`
}

type Task struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

type logger interface {
	Debugf(string, ...interface{})
}

func NewManager(token string) *Manager {
	return &Manager{
		Client:    http.DefaultClient,
		BaseURL:   DefaultBaseURL,
		Token:     token,
		UserAgent: "Rustack-go",
		ctx:       context.Background(),
	}
}

func (m *Manager) WithContext(ctx context.Context) *Manager {
	newManager := *m
	newManager.ctx = ctx
	return &newManager
}

func (m *Manager) Request(method string, path string, args interface{}, target interface{}) error {
	m.log("[rustack] %s %s", method, path)

	res, err := json.Marshal(args)
	if err != nil {
		return err
	}

	m.log("[rustack] Send %s", res)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest(method, request_url, bytes.NewReader(res))
	if err != nil {
		return errors.Wrapf(err, "Invalid %s request %s", method, request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))
	req.Header.Set("Content-Type", "application/json")

	req = req.WithContext(m.ctx)

	taskIds, err := m.do(req, request_url, target, res)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) Get(path string, args Arguments, target interface{}) error {
	m.log("[rustack] GET %s", path)

	params := args.ToURLValues()

	request_url, _ := url.JoinPath(m.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", request_url, params.Encode())

	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	req = req.WithContext(m.ctx)

	_, err = m.do(req, request_url, target, nil)
	return err
}

func (m *Manager) GetItems(path string, args Arguments, target interface{}) error {
	items := []byte{"["[0]}

	params := args.ToURLValues()

	page := 1
	for {
		params.Set("page", fmt.Sprint(page))

		m.log("[rustack] GET %s?%s", path, params.Encode())

		request_url, _ := url.JoinPath(m.BaseURL, path)
		urlWithParams := fmt.Sprintf("%s?%s", request_url, params.Encode())

		req, err := http.NewRequest("GET", urlWithParams, nil)
		if err != nil {
			return errors.Wrapf(err, "Invalid GET request %s", request_url)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

		req = req.WithContext(m.ctx)

		type tempStruct struct {
			Total int             `json:"total"`
			Limit int             `json:"limit"`
			Items json.RawMessage // To future unmarshalling
		}

		temp := new(tempStruct)

		_, err = m.do(req, request_url, temp, nil)
		if err != nil {
			break
		}

		lastByte := temp.Items[len(temp.Items)-1]
		if lastByte != 93 { // Means "]"
			return errors.New("Unmarshalling is not going well")
		}

		items = append(items, temp.Items[1:len(temp.Items)-1]...) // Cut 1st and last bytes
		items = append(items, ","[0])                             // Add comma

		page++
	}

	if len(items) > 0 {
		items = append(items[0:len(items)-1], "]"[0]) // Remove the last comma and add closed brace
	}

	err := json.Unmarshal(items, target)

	if err != nil {
		return errors.Wrapf(err, "JSON items decode failed on %s:", path)
	}

	return nil
}

func (m *Manager) GetSubItems(path string, args Arguments, target interface{}) error {

	m.log("[rustack] GET %s", path)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest("GET", request_url, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	req = req.WithContext(m.ctx)

	_, err = m.do(req, request_url, target, nil)
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Delete(path string, args Arguments, target interface{}) error {
	m.log("[rustack] DELETE %s", path)

	request_url, _ := url.JoinPath(m.BaseURL, path)

	req, err := http.NewRequest("DELETE", request_url, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid DELETE request %s", request_url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	taskIds, err := m.do(req, request_url, target, nil)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) WaitTask(taskId string) error {
	m.log("[rustack] Start waiting task %s...", taskId)

	path, _ := url.JoinPath("v1/job", taskId)
	start := time.Now()
	var task Task

	for {
		err := m.Get(path, Arguments{}, task)
		if err != nil {
			break
		}
		if task.Status == "error" {
			return errors.New(fmt.Sprintf("Task in error status, step: %s", task.Name))
		}

		if err := m.sleep(RetryTime * time.Millisecond); err != nil {
			return err
		}

		elapsedTime := time.Since(start)

		if elapsedTime.Seconds() > float64(TaskTimeout) {
			m.log("[rustack] Waiting task %s took more than %ds", taskId, TaskTimeout)
			return errors.New("Task timeout")
		}
	}

	m.log("[rustack] End waiting task %s", taskId)

	return nil
}

func (m *Manager) log(format string, args ...interface{}) {
	if m.Logger != nil {
		m.Logger.Debugf(format, args...)
	}
}

func (m *Manager) sleep(dur time.Duration) error {
	if m.ctx != nil {
		return SleepWithContext(m.ctx, dur)
	} else {
		time.Sleep(dur)
	}

	return nil
}

// TODO: добавить 10 минут таймаута
func (m *Manager) do(req *http.Request, url string, target interface{}, requestBody []byte) (string, error) {
	req.Header.Set("Accept-Language", "ru-ru")
	var locked_object ObjectLocked

	start := time.Now()
	var resp *http.Response

	for {
		m.log("[rustack] Perform %s...", req.Method)

		req.Body = ioutil.NopCloser(bytes.NewReader(requestBody))
		resp_, err := m.Client.Do(req)
		if err != nil {
			return "", errors.Wrapf(err, "HTTP request failure on %s", url)
		}

		defer resp_.Body.Close()

		if resp_.StatusCode == 409 {
			m.log("[rustack] Object '%s' locked. Try again in %dms...", url, RetryTime)
			body, err := ioutil.ReadAll(resp_.Body)
			err = json.Unmarshal(body, &locked_object)


			if err != nil {
				return "", errors.Wrapf(err, "HTTP Read error on response for %s", url)
			}
			
			if locked_object.ErrorAlias != nil {
				error_alias := fmt.Sprintf("%v", locked_object.ErrorAlias[0])
				error_details, _ := json.Marshal(locked_object.Details)
				error_data := fmt.Sprintf("%v", locked_object.NonFieldErrors[0])
				if error_alias == "limit_exceeded" || error_alias == "object_protected" {
					error_body := fmt.Sprintf("%s: %s", error_data, string(error_details))
					return "", errors.New(error_body)
				}
			}

			if err := m.sleep(RetryTime * time.Millisecond); err != nil {
				return "", err
			}

			elapsedTime := time.Since(start)

			if elapsedTime.Seconds() > float64(LockTimeout) {
				m.log("[rustack] Waiting unlock for '%s' took more than %ds", url, LockTimeout)
				return "", errors.New("Lock timeout")
			}

			continue // try again
		}

		resp = resp_
		break
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		m.log("[rustack] Error response %d on '%s'", resp.StatusCode, url)
		return "", NewRustackApiError(url, resp)
	} else {
		m.log("[rustack] Success response on '%s'", url)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "HTTP Read error on response for %s", url)
	}

	// task waiter
	taskIds := resp.Header.Get("X-Esu-Tasks")
	if taskIds != "" {
		m.log("[rustack] Tasks IDS: %s", taskIds)
	}

	if len(b) == 0 {
		return taskIds, nil
	}

	if target == nil {
		// Don't try to unmarshall in case target is nil
		return taskIds, nil
	}

	err = json.Unmarshal(b, target)

	if err != nil {
		return "", errors.Wrapf(err, "JSON decode failed on %s:\n%s", url, string(b))
	}

	return taskIds, nil
}

func (m *Manager) waitTasks(taskIds string) error {
	for _, taskId := range strings.Split(taskIds, ",") {
		taskId := strings.TrimSpace(taskId)
		if taskId == "" {
			continue
		}

		if err := m.WaitTask(taskId); err != nil {
			return err
		}
	}

	return nil
}
