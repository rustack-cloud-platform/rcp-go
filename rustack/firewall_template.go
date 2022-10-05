package rustack

import (
	"fmt"
	"net/url"
)

type FirewallTemplate struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Locked  bool   `json:"locked"`
}

func (m *Manager) GetFirewallTemplate(id string) (firewallTemplate *FirewallTemplate, err error) {
	path, _ := url.JoinPath("v1/firewall/", id)
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

func NewFirewallTemplate(name string) (firewallTemplate FirewallTemplate) {
	d := FirewallTemplate{Name: name}
	return d
}

func (f *FirewallTemplate) Update(firewallRule *FirewallRule) (err error) {

	path := fmt.Sprintf("v1/firewall/%s/rule", f.ID)

	err = f.manager.Request("POST", path, firewallRule, &firewallRule)
	if err == nil {
		firewallRule.manager = f.manager
	}
	return
}

func (f *FirewallTemplate) Delete() (err error) {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	return f.manager.Delete(path, Defaults(), &f)
}

func (f *FirewallTemplate) Rename(name string) (err error) {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	return f.manager.Request("PUT", path, Arguments{"name": name}, &f)
}

func (v *Vdc) CreateFirewallTemplate(firewallTemplate *FirewallTemplate) (err error) {
	args := &struct {
		Name string  `json:"name"`
		Vdc  *string `json:"vdc,omitempty"`
	}{
		Name: firewallTemplate.Name,
		Vdc:  &v.ID,
	}

	err = v.manager.Request("POST", "v1/firewall", args, &firewallTemplate)
	if err == nil {
		firewallTemplate.manager = v.manager
	}
	return
}

func (f FirewallTemplate) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/firewall", f.ID)
	return loopWaitLock(f.manager, path)
}
