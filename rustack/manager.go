package rustack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
	Logger    logger
	BaseURL   string
	Token     string
	UserAgent string
	ctx       context.Context
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

func (m *Manager) Get(path string, args Arguments, target interface{}) error {
	m.log("[rustack] GET %s", path)

	params := args.ToURLValues()

	url := fmt.Sprintf("%s/%s", m.BaseURL, path)
	urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid GET request %s", url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	req = req.WithContext(m.ctx)

	_, err = m.do(req, url, target)
	return err
}

func (m *Manager) GetItems(path string, args Arguments, target interface{}) error {
	items := []byte{"["[0]}

	params := args.ToURLValues()

	page := 1
	for {
		params.Set("page", fmt.Sprint(page))

		m.log("[rustack] GET %s?%s", path, params.Encode())

		url := fmt.Sprintf("%s/%s", m.BaseURL, path)
		urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

		req, err := http.NewRequest("GET", urlWithParams, nil)
		if err != nil {
			return errors.Wrapf(err, "Invalid GET request %s", url)
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

		req = req.WithContext(m.ctx)

		type tempStruct struct {
			Total int             `json:"total"`
			Limit int             `json:"limit"`
			Items json.RawMessage // To future unmarshalling
		}

		temp := new(tempStruct)

		_, err = m.do(req, url, temp)
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

func (m *Manager) Put(path string, args interface{}, target interface{}) error {
	m.log("[rustack] PUT %s", path)

	res, err := json.Marshal(args)
	m.log("[rustack] Send %s", res)

	url := fmt.Sprintf("%s/%s", m.BaseURL, path)
	// urlWithParams := fmt.Sprintf("%s?%s", url, params.Encode())

	req, err := http.NewRequest("PUT", url, bytes.NewReader(res))
	if err != nil {
		return errors.Wrapf(err, "Invalid PUT request %s", url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))
	req.Header.Set("Content-Type", "application/json")

	taskIds, err := m.do(req, url, target)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) Post(path string, args interface{}, target interface{}) error {
	m.log("[rustack] POST %s", path)

	res, err := json.Marshal(args)
	m.log("[rustack] Send %s", res)

	url := fmt.Sprintf("%s/%s", m.BaseURL, path)

	req, err := http.NewRequest("POST", url, bytes.NewReader(res))
	if err != nil {
		return errors.Wrapf(err, "Invalid POST request %s", url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))
	req.Header.Set("Content-Type", "application/json")

	taskIds, err := m.do(req, url, target)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) Delete(path string, args Arguments, target interface{}) error {
	m.log("[rustack] DELETE %s", path)

	url := fmt.Sprintf("%s/%s", m.BaseURL, path)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.Wrapf(err, "Invalid DELETE request %s", url)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", m.Token))

	taskIds, err := m.do(req, url, target)
	m.waitTasks(taskIds)

	return err
}

func (m *Manager) log(format string, args ...interface{}) {
	if m.Logger != nil {
		m.Logger.Debugf(format, args...)
	}
}

func (m *Manager) do(req *http.Request, url string, target interface{}) (string, error) {
	req.Header.Set("Accept-Language", "ru-ru")

	start := time.Now()
	var resp *http.Response

	for {
		m.log("[rustack] Perform '%s' to '%s'...", req.Method, url)

		resp_, err := m.Client.Do(req)
		if err != nil {
			return "", errors.Wrapf(err, "HTTP request failure on %s", url)
		}
		defer resp_.Body.Close()

		if resp_.StatusCode == 409 {
			m.log("[rustack] Object '%s' locked. Try again in %dms...", url, RetryTime)

			time.Sleep(RetryTime * time.Millisecond)

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
		return "", makeHTTPClientError(url, resp)
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

		m.log("[rustack] Start waiting task %s...", taskId)
		path := fmt.Sprintf("v1/job/%s", taskId)
		start := time.Now()

		for {
			err := m.Get(path, Arguments{}, nil)

			if err != nil {
				break
			}

			time.Sleep(RetryTime * time.Millisecond)

			elapsedTime := time.Since(start)

			if elapsedTime.Seconds() > float64(TaskTimeout) {
				m.log("[rustack] Waiting task %s took more than %ds", taskId, TaskTimeout)
				return errors.New("Task timeout")
			}
		}
		m.log("[rustack] End waiting task %s", taskId)
	}

	return nil
}
