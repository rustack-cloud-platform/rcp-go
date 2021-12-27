package rustack

import (
	"fmt"
)

type Vm struct {
	manager     *Manager
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Cpu         int           `json:"cpu"`
	Ram         int           `json:"ram"`
	Power       bool          `json:"power"`
	Vdc         *Vdc          `json:"vdc"`
	HotAdd      bool          `json:"hotadd_feature"`
	Template    *Template     `json:"template"`
	Metadata    []*VmMetadata `json:"metadata"`
	UserData    *string       `json:"user_data"`
	Ports       []*Port       `json:"ports"`
	Disks       []*Disk       `json:"disks"`
	Floating    *Port         `json:"floating"`
	Locked      bool          `json:"locked,omitempty"`
	Kubernetes  *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"kubernetes,omitempty"`
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
			vms[i].Disks[x].manager = m
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
	for x := range vm.Ports {
		vm.Ports[x].manager = m
	}
	for x := range vm.Disks {
		vm.Disks[x].manager = m
	}
	vm.Vdc.manager = m
	if vm.Floating != nil {
		vm.Floating.manager = m
	}
	return
}

func (v *Vm) Reload() error {
	m := v.manager
	path := fmt.Sprintf("v1/vm/%s", v.ID)
	if err := m.Get(path, Defaults(), &v); err != nil {
		return err
	}

	v.manager = m
	for x := range v.Ports {
		v.Ports[x].manager = m
	}
	for x := range v.Disks {
		v.Disks[x].manager = m
	}
	v.Vdc.manager = m
	if v.Floating != nil {
		v.Floating.manager = m
	}

	return nil
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

func (v *Vm) Update() error {
	path := fmt.Sprintf("v1/vm/%s", v.ID)
	args := &struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Cpu         int     `json:"cpu"`
		Ram         int     `json:"ram"`
		HotAdd      bool    `json:"hotadd_feature"`
		Floating    *string `json:"floating"`
	}{
		Name:        v.Name,
		Description: v.Description,
		Cpu:         v.Cpu,
		Ram:         v.Ram,
		HotAdd:      v.HotAdd,
		Floating:    nil,
	}

	if v.Floating != nil {
		if v.Floating.ID != "" {
			args.Floating = &v.Floating.ID
		} else {
			args.Floating = v.Floating.IpAddress
		}
	}

	return v.manager.Put(path, args, v)
}

func (v *Vm) updateState(state string) error {
	path := fmt.Sprintf("v1/vm/%s/state", v.ID)

	args := &struct {
		State string `json:"state"`
	}{
		State: state,
	}

	return v.manager.Post(path, args, v)
}

func (v *Vm) PowerOn() error {
	return v.updateState("power_on")
}

func (v *Vm) Reboot() error {
	return v.updateState("reboot")
}

func (v *Vm) PowerOff() error {
	return v.updateState("power_off")
}

func (v *Vm) Delete() error {
	path := fmt.Sprintf("v1/vm/%s", v.ID)
	return v.manager.Delete(path, Defaults(), v)
}

func (v Vm) WaitLock() (err error) {
	path := fmt.Sprintf("v1/vm/%s", v.ID)
	return loopWaitLock(v.manager, path)
}
