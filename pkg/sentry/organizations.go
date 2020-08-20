package sentry

import (
	"fmt"
	"net/http"
	"time"
)

type OrganizationsService service

type Organization struct {
	Avatar         Avatar             `json:"avatar"`
	DateCreated    time.Time          `json:"dateCreated"`
	ID             string             `json:"id"`
	IsEarlyAdopter bool               `json:"isEarlyAdopter"`
	Name           string             `json:"name"`
	Require2FA     bool               `json:"require2FA"`
	Slug           string             `json:"slug"`
	Status         OrganizationStatus `json:"status"`
}

type OrganizationStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (s *OrganizationsService) Get(organizationSlug string) (*Organization, *Response, error) {
	endpoint := fmt.Sprintf("/organizations/%s", organizationSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	organization := new(Organization)
	resp, err := s.client.do(req, organization)
	return organization, resp, err
}

func (s *OrganizationsService) ListProjects(organizationSlug string, opts *ListOptions) ([]Project, *Response, error) {
	var endpoint string
	if opts.Cursor == "" {
		endpoint = fmt.Sprintf("/organizations/%s/projects", organizationSlug)
	} else {
		endpoint = fmt.Sprintf("/organizations/%s/projects/?&cursor=%s", organizationSlug, opts.Cursor)
	}

	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	projects := new([]Project)
	resp, err := s.client.do(req, projects)
	return *projects, resp, err
}
