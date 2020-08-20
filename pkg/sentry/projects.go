package sentry

import (
	"fmt"
	"net/http"
	"time"
)

type ProjectsService service

type Project struct {
	Avatar       Avatar       `json:"avatar"`
	Color        string       `json:"color"`
	DateCreated  time.Time    `json:"dateCreated"`
	Features     []string     `json:"features"`
	FirstEvent   time.Time    `json:"firstEvent"`
	HasAccess    bool         `json:"hasAccess"`
	ID           string       `json:"id"`
	IsBookmarked bool         `json:"isBookmarked"`
	IsInternal   bool         `json:"isInternal"`
	IsMember     bool         `json:"isMember"`
	IsPublic     bool         `json:"isPublic"`
	Name         string       `json:"name"`
	Organization Organization `json:"organization"`
	Platform     string       `json:"platform"`
	Slug         string       `json:"slug"`
	Status       string       `json:"status"`
	Team         Team         `json:"team"`
	Teams        []Team       `json:"teams"`
}

func (s *ProjectsService) List(opts *ListOptions) ([]Project, *Response, error) {
	var endpoint string
	if opts.Cursor == "" {
		endpoint = "/projects"
	} else {
		endpoint = fmt.Sprintf("/projects/?&cursor=%s", opts.Cursor)
	}

	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	projects := new([]Project)
	resp, err := s.client.do(req, projects)
	return *projects, resp, err
}

func (s *ProjectsService) Get(organizationSlug, projectSlug string) (*Project, *Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s", organizationSlug, projectSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	project := new(Project)
	resp, err := s.client.do(req, project)
	return project, resp, err
}

type UpdateProjectParams struct {
	Name            string `json:"name,omitempty"`
	Slug            string `json:"slug,omitempty"`
	Team            string `json:"team,omitempty"`
	Platform        string `json:"platform,omitempty"`
	IsBookmarked    *bool  `json:"isBookmarked,omitempty"`
	DigestsMinDelay int    `json:"digestsMinDelay,omitempty"`
	DigestsMaxDelay int    `json:"digestsMaxDelay,omitempty"`
}

func (s *ProjectsService) Update(organizationSlug, projectSlug string, params *UpdateProjectParams) (*Project, *Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s", organizationSlug, projectSlug)
	req, err := s.client.newRequest(http.MethodPut, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	project := new(Project)
	resp, err := s.client.do(req, project)
	return project, resp, err
}

func (s *ProjectsService) Delete(organizationSlug, projectSlug string) (*Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s", organizationSlug, projectSlug)
	req, err := s.client.newRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(req, nil)
	return resp, err
}

type ProjectKey struct {
	BrowserSDK        ProjectKeyBrowserSDK `json:"browserSdk"`
	BrowserSDKVersion string               `json:"browserSdkVersion"`
	DateCreated       time.Time            `json:"dateCreated"`
	DSN               ProjectKeyDSN        `json:"dsn"`
	ID                string               `json:"id"`
	IsActive          bool                 `json:"isActive"`
	Label             string               `json:"label"`
	Name              string               `json:"name"`
	ProjectID         int                  `json:"projectId"`
	Public            string               `json:"public"`
	RateLimit         ProjectKeyRateLimit  `json:"rateLimit"`
	Secret            string               `json:"secret"`
}

type ProjectKeyBrowserSDK struct {
	Choices [][]string `json:"choices"`
}

type ProjectKeyDSN struct {
	CDN      string `json:"cdn"`
	CSP      string `json:"csp"`
	Minidump string `json:"minidump"`
	Public   string `json:"public"`
	Secret   string `json:"secret"`
	Security string `json:"security"`
}

type ProjectKeyRateLimit struct {
	Window int `json:"window"`
	Count  int `json:"count"`
}

func (s *ProjectsService) ListKeys(organizationSlug, projectSlug string, opts *ListOptions) ([]ProjectKey, *Response, error) {
	var endpoint string
	if opts.Cursor == "" {
		endpoint = fmt.Sprintf("/projects/%s/%s/keys", organizationSlug, projectSlug)
	} else {
		endpoint = fmt.Sprintf("/projects/%s/%s/keys/?&cursor=%s", organizationSlug, projectSlug, opts.Cursor)
	}

	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	keys := new([]ProjectKey)
	resp, err := s.client.do(req, keys)
	return *keys, resp, err
}

func (s *ProjectsService) GetKey(organizationSlug, projectSlug, keyID string) (*ProjectKey, *Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s/keys/%s", organizationSlug, projectSlug, keyID)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	key := new(ProjectKey)
	resp, err := s.client.do(req, key)
	return key, resp, err
}

type CreateProjectKeyParams struct {
	Name string `json:"name,omitempty"`
}

func (s *ProjectsService) CreateKey(organizationSlug, projectSlug string, params *CreateProjectKeyParams) (*ProjectKey, *Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s/keys", organizationSlug, projectSlug)
	req, err := s.client.newRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	key := new(ProjectKey)
	resp, err := s.client.do(req, key)
	return key, resp, err
}

type UpdateProjectKeyParams struct {
	Name string `json:"name,omitempty"`
}

func (s *ProjectsService) UpdateKey(organizationSlug, projectSlug, keyID string, params *UpdateProjectKeyParams) (*ProjectKey, *Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s/keys/%s", organizationSlug, projectSlug, keyID)
	req, err := s.client.newRequest(http.MethodPut, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	key := new(ProjectKey)
	resp, err := s.client.do(req, key)
	return key, resp, err
}

func (s *ProjectsService) DeleteKey(organizationSlug, projectSlug, keyID string) (*Response, error) {
	endpoint := fmt.Sprintf("/projects/%s/%s/keys/%s", organizationSlug, projectSlug, keyID)
	req, err := s.client.newRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(req, nil)
	return resp, err
}
