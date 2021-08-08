package rustack

type FirewallTemplate struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func (v *Vdc) GetFirewallTemplate() (firewallTemplate []*FirewallTemplate, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	path := "v1/firewall"
	err = v.manager.GetItems(path, args, &firewallTemplate)
	for i := range firewallTemplate {
		firewallTemplate[i].manager = v.manager
	}
	return
}
