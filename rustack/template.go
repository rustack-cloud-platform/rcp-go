package rustack

type Template struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	MinCpu  int    `json:"min_cpu"`
	MinRam  int    `json:"min_ram"`
	MinHdd  int    `json:"min_hdd"`
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
