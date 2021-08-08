package rustack

import "fmt"

type Project struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Client  struct {
		Id string `json:"id"`
	} `json:"client"`
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
	path := fmt.Sprintf("v1/project/%s", id)
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

	err := c.manager.Post("v1/project", args, &project)
	if err == nil {
		project.manager = c.manager
	}

	return err
}

func (p *Project) Rename(name string) error {
	path := fmt.Sprintf("v1/project/%s", p.ID)
	return p.manager.Put(path, Arguments{"name": name, "client": p.Client.Id}, p)
}

func (p *Project) Delete() error {
	path := fmt.Sprintf("v1/project/%s", p.ID)
	return p.manager.Delete(path, Defaults(), p)
}
