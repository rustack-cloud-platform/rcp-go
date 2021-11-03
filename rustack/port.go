package rustack

import "fmt"

type Port struct {
	manager           *Manager
	ID                string              `json:"id"`
	IpAddress         *string             `json:"ip_address,omitempty"`
	Network           *Network            `json:"network"`
	FirewallTemplates []*FirewallTemplate `json:"fw_templates"`
}

func NewPort(network *Network, firewallTemplates []*FirewallTemplate, ipAddress *string) Port {
	p := Port{Network: network, FirewallTemplates: firewallTemplates, IpAddress: ipAddress}
	return p
}

func (v *Vdc) GetPorts(extraArgs ...Arguments) (ports []*Port, err error) {
	args := Arguments{
		"vdc": v.ID,
	}

	args.merge(extraArgs)

	path := "v1/port"
	err = v.manager.GetItems(path, args, &ports)
	for i := range ports {
		ports[i].manager = v.manager
	}
	return
}

func (p *Port) UpdateFirewall(firewallTemplates []*FirewallTemplate) error {
	path := fmt.Sprintf("v1/port/%s", p.ID)

	var fwTemplates = make([]*string, 0)
	for _, fwTemplate := range firewallTemplates {
		fwTemplates = append(fwTemplates, &fwTemplate.ID)
	}

	args := &struct {
		FwTemplates []*string `json:"fw_templates"`
	}{
		FwTemplates: fwTemplates,
	}

	err := p.manager.Put(path, args, nil)
	if err != nil {
		return err
	}

	return nil
}

func (p *Port) Delete() error {
	path := fmt.Sprintf("v1/port/%s", p.ID)
	return p.manager.Delete(path, Defaults(), p)
}
