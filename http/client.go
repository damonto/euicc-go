package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	Client               *http.Client
	AdminProtocolVersion string
}

func (c *Client) NewRequest(u *url.URL, request any) (*http.Request, error) {
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(request); err != nil {
		return nil, err
	}
	httpRequest, err := http.NewRequest(http.MethodPost, u.String(), &body)
	if err != nil {
		return nil, err
	}
	httpRequest.Header = c.Header()
	return httpRequest, nil
}

func (c *Client) SendRequest(u *url.URL, request, response any) error {
	httpRequest, err := c.NewRequest(u, request)
	if err != nil {
		return err
	}
	httpResponse, err := c.Client.Do(httpRequest)
	if err != nil {
		return err
	}
	defer httpResponse.Body.Close()
	if httpResponse.StatusCode > 299 {
		return fmt.Errorf("unexpected status code: %d", httpResponse.StatusCode)
	}
	if err = json.NewDecoder(httpResponse.Body).Decode(response); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (c *Client) Header() http.Header {
	return http.Header{
		"Content-Type":     {"application/json"},
		"User-Agent":       {"gsma-rsp-lpad"},
		"X-Admin-Protocol": {fmt.Sprintf("gsma/rsp/v%s", c.AdminProtocolVersion)},
	}
}
