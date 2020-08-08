package sentry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	DefaultSentryURL = "https://sentry.io/"
	APIVersion       = 0
)

type Client struct {
	client  *http.Client
	token   string
	baseURL *url.URL

	Organizations *OrganizationsService
	Projects      *ProjectsService
	Teams         *TeamsService
}

type ClientOption func(*Client)

type service struct {
	client *Client
}

func NewClient(token string, opts ...ClientOption) *Client {
	sentryURL, _ := url.Parse(DefaultSentryURL)
	sentryURL.Path = path.Join("api", strconv.Itoa(APIVersion)) + "/"

	client := &Client{
		client:  &http.Client{},
		baseURL: sentryURL,
		token:   token,
	}

	for _, opt := range opts {
		opt(client)
	}

	common := service{client}
	client.Organizations = (*OrganizationsService)(&common)
	client.Projects = (*ProjectsService)(&common)
	client.Teams = (*TeamsService)(&common)

	return client
}

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.client = httpClient
	}
}

func WithSentryURL(sentryURL *url.URL) ClientOption {
	return func(c *Client) {
		sentryURL.Path = path.Join("api", strconv.Itoa(APIVersion)) + "/"
		c.baseURL = sentryURL
	}
}

func (c *Client) do(req *http.Request, v interface{}) (*Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	response := c.newResponse(resp)

	if code := response.StatusCode; code < 200 || code > 299 {
		apiErr := make(APIError)
		err := json.NewDecoder(response.Body).Decode(&apiErr)
		if err != nil {
			return nil, err
		}

		return response, apiErr
	}

	if v != nil {
		err = json.NewDecoder(response.Body).Decode(v)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return response, nil
}

func (c *Client) newRequest(method, endpoint string, body interface{}) (*http.Request, error) {
	endpoint = strings.Trim(endpoint, "/") + "/"
	requestURL, err := c.baseURL.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, requestURL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	return req, nil
}

type Response struct {
	*http.Response

	PrevPage *Page
	NextPage *Page
}

func (c *Client) newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	response.parsePaginationLinks()
	return response
}
