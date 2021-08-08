package rustack

import "fmt"

type Vm struct {
	manager  *Manager
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Cpu      int           `json:"cpu"`
	Ram      int           `json:"ram"`
	Power    bool          `json:"power"`
	Vdc      *Vdc          `json:"vdc"`
	Template *Template     `json:"template"`
	Metadata []*VmMetadata `json:"metadata"`
	UserData *string       `json:"user_data"`
	Ports    []*Port       `json:"ports"`
	Disks    []*Disk       `json:"disks"`
	Floating *Port         `json:"floating"`
}

func NewVm(name string, cpu, ram int, template *Template, metadata []*VmMetadata, userData *string, ports []*Port, disks []*Disk, floating *string) Vm {
	v := Vm{Name: name, Cpu: cpu, Ram: ram, Power: true, Template: template, Metadata: metadata, UserData: userData, Ports: ports, Disks: disks}
	if floating != nil {
		v.Floating = &Port{IpAddress: floating}
	}
	return v
}

func (m *Manager) GetVms(extraArgs ...Arguments) (vms []*Vm, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/vm"
	err = m.GetItems(path, args, &vms)
	for i := range vms {
		vms[i].manager = m
		for x := range vms[i].Ports {
			vms[i].Ports[x].manager = m
		}
		for x := range vms[i].Disks {
			vms[i].Ports[x].manager = m
		}
		vms[i].Vdc.manager = m
		if vms[i].Floating != nil {
			vms[i].Floating.manager = m
		}
	}
	return
}

func (v *Vdc) GetVms(extraArgs ...Arguments) (vms []*Vm, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	vms, err = v.manager.GetVms(args)
	return
}

func (m *Manager) GetVm(id string) (vm *Vm, err error) {
	path := fmt.Sprintf("v1/vm/%s", id)
	err = m.Get(path, Defaults(), &vm)
	if err != nil {
		return
	}
	vm.manager = m
	return
}

func (v *Vm) AddPort(port *Port) error {
	type TempPortCreate struct {
		Vm          string   `json:"vm"`
		Network     string   `json:"network"`
		IpAddress   *string  `json:"ip_address,omitempty"`
		FwTemplates []string `json:"fw_templates"`
	}

	var fwTemplates = make([]string, len(port.FirewallTemplates))
	for i, fwTemplate := range port.FirewallTemplates {
		fwTemplates[i] = fwTemplate.ID
	}
	args := &TempPortCreate{
		Vm:          v.ID,
		Network:     port.Network.ID,
		IpAddress:   port.IpAddress,
		FwTemplates: fwTemplates,
	}

	err := v.manager.Post("v1/port", args, &port)
	if err == nil {
		port.manager = v.manager
	}

	return err
}

func (v *Vm) Delete() error {
	path := fmt.Sprintf("v1/vm/%s", v.ID)
	return v.manager.Delete(path, Defaults(), v)
}
