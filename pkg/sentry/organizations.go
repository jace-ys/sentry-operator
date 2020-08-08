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

func (s *OrganizationsService) ListProjects(organizationSlug string) ([]Project, *Response, error) {
	endpoint := fmt.Sprintf("/organizations/%s/projects", organizationSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	projects := new([]Project)
	resp, err := s.client.do(req, projects)
	return *projects, resp, err
}
