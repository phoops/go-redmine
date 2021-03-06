package redmine

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type Client struct {
	endpoint   string
	apikey     string
	switchUser string
	*http.Client
	Limit  int
	Offset int
}

// FullTraversingClient is a Client that automatically traverse pagination
type FullTraversingClient struct {
	*Client
}

var DefaultLimit int = -1  // "-1" means "No setting"
var DefaultOffset int = -1 //"-1" means "No setting"

func NewClient(endpoint, apikey string) *Client {
	return &Client{endpoint, apikey, "", http.DefaultClient, DefaultLimit, DefaultOffset}
}

func NewFullTraversingClient(endpoint, apikey string) *FullTraversingClient {
	//return &FullTraversingClient{NewClient(endpoint, apikey)}
	return &FullTraversingClient{&Client{endpoint, apikey, "", http.DefaultClient, 100, 0}}
}

// URLWithFilter return string url by concat endpoint, path and filter
// err != nil when endpoin can not parse
func (c *Client) URLWithFilter(path string, f Filter) (string, error) {
	var fullURL *url.URL
	fullURL, err := url.Parse(c.endpoint)
	if err != nil {
		return "", err
	}
	fullURL.Path += path
	if c.Limit > -1 {
		f.AddPair("limit", strconv.Itoa(c.Limit))
	}
	if c.Offset > -1 {
		f.AddPair("offset", strconv.Itoa(c.Offset))
	}
	fullURL.RawQuery = f.ToURLParams()
	return fullURL.String(), nil
}

func (c *Client) getPaginationClause() string {
	clause := ""
	if c.Limit > -1 {
		clause = clause + fmt.Sprintf("&limit=%v", c.Limit)
	}
	if c.Offset > -1 {
		clause = clause + fmt.Sprintf("&offset=%v", c.Offset)
	}
	return clause
}

func (c *Client) SwitchUser(userLogin string) {
	c.switchUser = userLogin
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("X-Redmine-API-Key", c.apikey)
	if c.switchUser != "" {
		req.Header.Add("X-Redmine-Switch-User", c.switchUser)
	}
	return c.Client.Do(req)
}

type errorsResult struct {
	Errors []string `json:"errors"`
}

type IdName struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Id struct {
	Id int `json:"id"`
}

type pagination struct {
	TotalCount int `json:"total_count"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
}
