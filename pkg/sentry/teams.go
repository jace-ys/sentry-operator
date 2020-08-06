package sentry

import (
	"fmt"
	"net/http"
	"time"
)

type TeamsService service

type Team struct {
	Avatar      Avatar    `json:"avatar"`
	DateCreated time.Time `json:"dateCreated"`
	HasAccess   bool      `json:"hasAccess"`
	ID          string    `json:"id"`
	IsMember    bool      `json:"isMember"`
	IsPending   bool      `json:"isPending"`
	MemberCount int       `json:"memberCount"`
	Name        string    `json:"name"`
	Projects    []Project `json:"projects"`
	Slug        string    `json:"slug"`
}

func (s *TeamsService) List(organizationSlug string) ([]Team, *Response, error) {
	endpoint := fmt.Sprintf("/organizations/%s/teams", organizationSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	teams := new([]Team)
	resp, err := s.client.do(req, teams)
	return *teams, resp, err
}

func (s *TeamsService) Get(organizationSlug, teamSlug string) (*Team, *Response, error) {
	endpoint := fmt.Sprintf("/teams/%s/%s", organizationSlug, teamSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	team := new(Team)
	resp, err := s.client.do(req, team)
	return team, resp, err
}

type CreateTeamParams struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

func (s *TeamsService) Create(organizationSlug string, params *CreateTeamParams) (*Team, *Response, error) {
	endpoint := fmt.Sprintf("/organizations/%s/teams", organizationSlug)
	req, err := s.client.newRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	team := new(Team)
	resp, err := s.client.do(req, team)
	return team, resp, err
}

type UpdateTeamParams struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

func (s *TeamsService) Update(organizationSlug, teamSlug string, params *UpdateTeamParams) (*Team, *Response, error) {
	endpoint := fmt.Sprintf("/teams/%s/%s", organizationSlug, teamSlug)
	req, err := s.client.newRequest(http.MethodPut, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	team := new(Team)
	resp, err := s.client.do(req, team)
	return team, resp, err
}

func (s *TeamsService) Delete(organizationSlug, teamSlug string) (*Response, error) {
	endpoint := fmt.Sprintf("/teams/%s/%s", organizationSlug, teamSlug)
	req, err := s.client.newRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.do(req, nil)
	return resp, err
}

func (s *TeamsService) ListProjects(organizationSlug, teamSlug string) ([]Project, *Response, error) {
	endpoint := fmt.Sprintf("/teams/%s/%s/projects", organizationSlug, teamSlug)
	req, err := s.client.newRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, nil, err
	}

	projects := new([]Project)
	resp, err := s.client.do(req, projects)
	return *projects, resp, err
}

type CreateProjectParams struct {
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

func (s *TeamsService) CreateProject(organizationSlug, teamSlug string, params *CreateProjectParams) (*Project, *Response, error) {
	endpoint := fmt.Sprintf("/teams/%s/%s/projects", organizationSlug, teamSlug)
	req, err := s.client.newRequest(http.MethodPost, endpoint, params)
	if err != nil {
		return nil, nil, err
	}

	project := new(Project)
	resp, err := s.client.do(req, project)
	return project, resp, err
}
