package sentry

import (
	"fmt"
	"net/http"
	"time"
)

type ProjectsService service

type Project struct {
	Avatar       Avatar    `json:"avatar"`
	Color        string    `json:"color"`
	DateCreated  time.Time `json:"dateCreated"`
	Features     []string  `json:"features"`
	FirstEvent   time.Time `json:"firstEvent"`
	HasAccess    bool      `json:"hasAccess"`
	ID           string    `json:"id"`
	IsBookmarked bool      `json:"isBookmarked"`
	IsInternal   bool      `json:"isInternal"`
	IsMember     bool      `json:"isMember"`
	IsPublic     bool      `json:"isPublic"`
	Name         string    `json:"name"`
	Platform     string    `json:"platform"`
	Slug         string    `json:"slug"`
	Status       string    `json:"status"`
	Team         Team      `json:"team"`
	Teams        []Team    `json:"teams"`
}

func (s *ProjectsService) List() ([]Project, *Response, error) {
	req, err := s.client.newRequest(http.MethodGet, "/projects", nil)
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
