package moodle

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"student-kanban/internal/domain/entity"
	"time"
)

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type SiteInfo struct {
	UserID         int64  `json:"userid"`
	Username       string `json:"username"`
	FullName       string `json:"fullname"`
	SiteName       string `json:"sitename"`
	SiteURL        string `json:"siteurl"`
	UserPictureURL string `json:"userpictureurl,omitempty"`
}

type Course struct {
	ID          int64  `json:"id"`
	ShortName   string `json:"shortname"`
	FullName    string `json:"fullname"`
	DisplayName string `json:"displayname"`
	Summary     string `json:"summary"`
	ViewURL     string `json:"viewurl"`
}

type CourseSection struct {
	ID      int64          `json:"id"`
	Name    string         `json:"name"`
	Modules []CourseModule `json:"modules"`
}

type CourseModule struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	ModName     string             `json:"modname"`
	URL         string             `json:"url"`
	Description string             `json:"description"`
	UserVisible bool               `json:"uservisible"`
	Dates       []CourseModuleDate `json:"dates"`
}

type CourseModuleDate struct {
	Label     string `json:"label"`
	Timestamp int64  `json:"timestamp"`
}

type tokenResponse struct {
	Token     string `json:"token"`
	Error     string `json:"error"`
	ErrorCode string `json:"errorcode"`
}

type moodleErrorResponse struct {
	Exception string `json:"exception"`
	ErrorCode string `json:"errorcode"`
	Message   string `json:"message"`
}

func NormalizeBaseURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("base URL is required")
	}

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", errors.New("invalid base URL")
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func (c *Client) ExchangeCredentialsForToken(baseURL, username, password, service string) (string, error) {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("service", service)

	endpoint := fmt.Sprintf("%s/login/token.php?%s", strings.TrimRight(baseURL, "/"), form.Encode())
	respBody, err := c.get(endpoint)
	if err != nil {
		return "", err
	}

	var response tokenResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	if response.Error != "" || response.Token == "" {
		return "", entity.ErrMoodleAuthFailed
	}

	return response.Token, nil
}

func (c *Client) GetSiteInfo(baseURL, token string) (*SiteInfo, error) {
	params := url.Values{}
	params.Set("wstoken", token)
	params.Set("wsfunction", "core_webservice_get_site_info")
	params.Set("moodlewsrestformat", "json")

	endpoint := fmt.Sprintf("%s/webservice/rest/server.php?%s", strings.TrimRight(baseURL, "/"), params.Encode())
	respBody, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var moodleErr moodleErrorResponse
	if err := json.Unmarshal(respBody, &moodleErr); err == nil && moodleErr.ErrorCode != "" {
		return nil, fmt.Errorf("moodle error: %s", moodleErr.Message)
	}

	var info SiteInfo
	if err := json.Unmarshal(respBody, &info); err != nil {
		return nil, fmt.Errorf("decode site info response: %w", err)
	}

	return &info, nil
}

func (c *Client) GetUserCourses(baseURL, token string, moodleUserID int64) ([]Course, error) {
	params := url.Values{}
	params.Set("wstoken", token)
	params.Set("wsfunction", "core_enrol_get_users_courses")
	params.Set("moodlewsrestformat", "json")
	params.Set("userid", strconv.FormatInt(moodleUserID, 10))
	params.Set("returnusercount", "0")

	endpoint := fmt.Sprintf("%s/webservice/rest/server.php?%s", strings.TrimRight(baseURL, "/"), params.Encode())
	respBody, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var moodleErr moodleErrorResponse
	if err := json.Unmarshal(respBody, &moodleErr); err == nil && moodleErr.ErrorCode != "" {
		return nil, fmt.Errorf("moodle error: %s", moodleErr.Message)
	}

	var courses []Course
	if err := json.Unmarshal(respBody, &courses); err != nil {
		return nil, fmt.Errorf("decode courses response: %w", err)
	}

	return courses, nil
}

func (c *Client) GetCourseContents(baseURL, token string, courseID int64) ([]CourseSection, error) {
	params := url.Values{}
	params.Set("wstoken", token)
	params.Set("wsfunction", "core_course_get_contents")
	params.Set("moodlewsrestformat", "json")
	params.Set("courseid", strconv.FormatInt(courseID, 10))

	endpoint := fmt.Sprintf("%s/webservice/rest/server.php?%s", strings.TrimRight(baseURL, "/"), params.Encode())
	respBody, err := c.get(endpoint)
	if err != nil {
		return nil, err
	}

	var moodleErr moodleErrorResponse
	if err := json.Unmarshal(respBody, &moodleErr); err == nil && moodleErr.ErrorCode != "" {
		return nil, fmt.Errorf("moodle error: %s", moodleErr.Message)
	}

	var sections []CourseSection
	if err := json.Unmarshal(respBody, &sections); err != nil {
		return nil, fmt.Errorf("decode course contents response: %w", err)
	}

	return sections, nil
}

func (c *Client) get(endpoint string) ([]byte, error) {
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= http.StatusBadRequest {
		message := strings.TrimSpace(string(body))
		if message == "" {
			return nil, fmt.Errorf("moodle request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("moodle request failed with status %d: %s", resp.StatusCode, message)
	}

	return body, nil
}
