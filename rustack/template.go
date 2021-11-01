package rustack

import "fmt"

type Template struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	MinCpu  int    `json:"min_cpu"`
	MinRam  int    `json:"min_ram"`
	MinHdd  int    `json:"min_hdd"`
}

func (m *Manager) GetTemplate(id string) (template *Template, err error) {
	path := fmt.Sprintf("v1/template/%s", id)
	err = m.Get(path, Defaults(), &template)
	if err != nil {
		return
	}
	template.manager = m
	return
}

func (v *Vdc) GetTemplates() (templates []*Template, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	path := "v1/template"
	err = v.manager.Get(path, args, &templates)
	for i := range templates {
		templates[i].manager = v.manager
	}
	return
}
