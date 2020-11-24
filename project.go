package redmine

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type projectRequest struct {
	Project Project `json:"project"`
}

type projectResult struct {
	Project Project `json:"project"`
}

type projectsResult struct {
	Projects []Project `json:"projects"`
	pagination
}

type Project struct {
	Id           int            `json:"id"`
	Parent       IdName         `json:"parent"`
	Name         string         `json:"name"`
	Identifier   string         `json:"identifier"`
	Description  string         `json:"description"`
	CreatedOn    string         `json:"created_on"`
	UpdatedOn    string         `json:"updated_on"`
	CustomFields []*CustomField `json:"custom_fields,omitempty"`
}

type ProjectsFilter struct {
	Filter
}

func NewProjectsFilter() *ProjectsFilter {
	return &ProjectsFilter{Filter{}}
}

const (
	ProjectStatusAll      string = ""
	ProjectStatusActive   string = "1"
	ProjectStatusClosed   string = "5"
	ProjectStatusArchived string = "9"
)

func (psf *ProjectsFilter) Status(status string) {
	psf.AddPair("status", status)
}

func (psf *ProjectsFilter) StatusNot(status string) {
	psf.AddPair("status", "!"+status)
}

func (c *Client) Project(id int) (*Project, error) {
	res, err := c.Get(c.endpoint + "/projects/" + strconv.Itoa(id) + ".json?key=" + c.apikey)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r projectResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r.Project, nil
}

func (c *Client) getProjects(filter *ProjectsFilter) (*projectsResult, error) {
	uri, err := c.URLWithFilter("/projects.json", filter.Filter)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Redmine-API-Key", c.apikey)
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r projectsResult
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *Client) Projects() ([]Project, error) {
	pr, err := c.getProjects(NewProjectsFilter())
	if err != nil {
		return nil, err
	}
	return pr.Projects, nil
}

func (c *Client) ProjectsWithFilter(filter *ProjectsFilter) ([]Project, error) {
	pr, err := c.getProjects(filter)
	if err != nil {
		return nil, err
	}
	return pr.Projects, nil
}

func (tc *FullTraversingClient) Projects() ([]Project, error) {
	return tc.ProjectsWithFilter(NewProjectsFilter())
}

func (tc *FullTraversingClient) ProjectsWithFilter(filter *ProjectsFilter) ([]Project, error) {
	curPage := 0
	maxPage := 1
	tc.Offset = 0
	projects := make([]Project, 0, 0)
	for proceed := true; proceed; proceed = curPage < maxPage {
		pr, err := tc.getProjects(filter)
		if err != nil {
			return nil, err
		}
		projects = append(projects, pr.Projects...)
		curPage++
		maxPage = 1 + (pr.TotalCount-1)/tc.Limit
		tc.Offset += tc.Limit
	}
	return projects, nil
}

func (c *Client) CreateProject(project Project) (*Project, error) {
	var ir projectRequest
	ir.Project = project
	s, err := json.Marshal(ir)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.endpoint+"/projects.json?key="+c.apikey, strings.NewReader(string(s)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var r projectRequest
	if res.StatusCode != 201 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	} else {
		err = decoder.Decode(&r)
	}
	if err != nil {
		return nil, err
	}
	return &r.Project, nil
}

func (c *Client) UpdateProject(project Project) error {
	var ir projectRequest
	ir.Project = project
	s, err := json.Marshal(ir)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", c.endpoint+"/projects/"+strconv.Itoa(project.Id)+".json?key="+c.apikey, strings.NewReader(string(s)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return errors.New("Not Found")
	}
	if res.StatusCode != 200 {
		decoder := json.NewDecoder(res.Body)
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	}
	if err != nil {
		return err
	}
	return err
}

func (c *Client) DeleteProject(id int) error {
	req, err := http.NewRequest("DELETE", c.endpoint+"/projects/"+strconv.Itoa(id)+".json?key="+c.apikey, strings.NewReader(""))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return errors.New("Not Found")
	}

	decoder := json.NewDecoder(res.Body)
	if res.StatusCode != 200 {
		var er errorsResult
		err = decoder.Decode(&er)
		if err == nil {
			err = errors.New(strings.Join(er.Errors, "\n"))
		}
	}
	return err
}
