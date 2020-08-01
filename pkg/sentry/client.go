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

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if code := resp.StatusCode; code < 200 || code > 299 {
		apiErr := make(APIError)
		err := json.NewDecoder(resp.Body).Decode(&apiErr)
		if err != nil {
			return nil, err
		}

		return resp, apiErr
	}

	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
	}

	return resp, nil
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
