package rustack

import "fmt"

type FirewallTemplate struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
}

func (m *Manager) GetFirewallTemplate(id string) (firewallTemplate *FirewallTemplate, err error) {
	path := fmt.Sprintf("v1/firewall/%s", id)
	err = m.Get(path, Defaults(), &firewallTemplate)
	if err != nil {
		return
	}
	firewallTemplate.manager = m
	return
}

func (v *Vdc) GetFirewallTemplates() (firewallTemplate []*FirewallTemplate, err error) {
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
