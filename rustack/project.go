package rustack

import (
	"net/url"
)

type Project struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Client  struct {
		Id string `json:"id"`
	} `json:"client"`
	Locked bool `json:"locked"`
}

func NewProject(name string) Project {
	b := Project{Name: name}
	return b
}

func (m *Manager) GetProjects(extraArgs ...Arguments) (projects []*Project, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/project"
	err = m.GetItems(path, args, &projects)
	for i := range projects {
		projects[i].manager = m
	}
	return
}

func (m *Manager) GetProject(id string) (project *Project, err error) {
	path, _ := url.JoinPath("v1/project", id)
	err = m.Get(path, Defaults(), &project)
	if err != nil {
		return
	}
	project.manager = m
	return
}

func (c *Client) CreateProject(project *Project) error {
	args := Arguments{
		"name":   project.Name,
		"client": c.ID,
	}

	err := c.manager.Request("POST", "v1/project", args, &project)
	if err == nil {
		project.manager = c.manager
	}

	return err
}

func (p *Project) Rename(name string) error {
	path, _ := url.JoinPath("v1/project", p.ID)
	return p.manager.Request("PUT", path, Arguments{"name": name, "client": p.Client.Id}, p)
}

func (p *Project) Delete() error {
	path, _ := url.JoinPath("v1/project", p.ID)
	return p.manager.Delete(path, Defaults(), nil)
}

func (p Project) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/project", p.ID)
	return loopWaitLock(p.manager, path)
}
