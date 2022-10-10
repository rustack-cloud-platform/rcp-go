package rustack

import (
	"net/url"
)

type Vdc struct {
	manager    *Manager
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Locked     bool       `json:"locked"`
	Hypervisor Hypervisor `json:"hypervisor"`

	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
}

func NewVdc(name string, hypervisor *Hypervisor) Vdc {
	v := Vdc{Name: name, Hypervisor: Hypervisor{ID: hypervisor.ID}}
	return v
}

func (m *Manager) GetVdcs(extraArgs ...Arguments) (vdcs []*Vdc, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/vdc"
	err = m.GetItems(path, args, &vdcs)
	for i := range vdcs {
		vdcs[i].manager = m
	}
	return
}

func (v *Vdc) GetVdcs(extraArgs ...Arguments) (vdcs []*Vdc, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	vdcs, err = v.manager.GetVdcs(args)
	return
}

func (m *Manager) GetVdc(id string) (vdc *Vdc, err error) {
	path, _ := url.JoinPath("v1/vdc", id)
	err = m.Get(path, Defaults(), &vdc)
	if err != nil {
		return
	}
	vdc.manager = m
	return
}

func (p *Project) CreateVdc(vdc *Vdc) error {
	args := Arguments{
		"name":       vdc.Name,
		"hypervisor": vdc.Hypervisor.ID,
		"project":    p.ID,
	}

	err := p.manager.Request("POST", "v1/vdc", args, &vdc)
	if err == nil {
		vdc.manager = p.manager
	}

	return err
}

func (v *Vdc) Rename(name string) error {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	return v.manager.Request("PUT", path, Arguments{"name": name}, v)
}

func (v *Vdc) Delete() error {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	return v.manager.Delete(path, Defaults(), v)
}

func (v *Vdc) CreateNetwork(network *Network) error {
	args := Arguments{
		"name": network.Name,
		"vdc":  v.ID,
	}

	err := v.manager.Request("POST", "v1/network", args, &network)
	if err == nil {
		network.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateRouter(router *Router, ports ...*Port) error {
	type TempPortCreate struct {
		ID string `json:"id"`
	}

	tempPorts := make([]*TempPortCreate, len(ports))
	for idx := range ports {
		tempPorts[idx] = &TempPortCreate{ID: ports[idx].ID}
	}

	args := &struct {
		Name     string            `json:"name"`
		Vdc      string            `json:"vdc"`
		Ports    []*TempPortCreate `json:"ports"`
		Floating *string           `json:"floating"`
	}{
		Name:     router.Name,
		Vdc:      v.ID,
		Ports:    tempPorts,
		Floating: nil,
	}

	if router.Floating != nil {
		if router.Floating.ID != "" {
			args.Floating = &router.Floating.ID
		} else {
			args.Floating = router.Floating.IpAddress
		}
	}

	err := v.manager.Request("POST", "v1/router", args, &router)
	if err == nil {
		router.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateVm(vm *Vm) error {
	type TempPortCreate struct {
		ID string `json:"id"`
	}

	tempPorts := make([]*TempPortCreate, len(vm.Ports))
	for idx := range vm.Ports {
		tempPorts[idx] = &TempPortCreate{ID: vm.Ports[idx].ID}
	}

	type TempFields struct {
		Field string `json:"field"`
		Value string `json:"value"`
	}

	tempFields := make([]*TempFields, len(vm.Metadata))
	for idx := range vm.Metadata {
		tempFields[idx] = &TempFields{Field: vm.Metadata[idx].Field.ID, Value: vm.Metadata[idx].Value}
	}

	type TempDisk struct {
		Name           string `json:"name"`
		Size           int    `json:"size"`
		StorageProfile string `json:"storage_profile"`
	}

	tempDisks := make([]*TempDisk, len(vm.Disks))
	for idx := range vm.Disks {
		tempDisks[idx] = &TempDisk{Name: vm.Disks[idx].Name, Size: vm.Disks[idx].Size, StorageProfile: vm.Disks[idx].StorageProfile.ID}
	}

	args := &struct {
		Name     string            `json:"name"`
		Cpu      int               `json:"cpu"`
		Ram      int               `json:"ram"`
		Vdc      string            `json:"vdc"`
		Template string            `json:"template"`
		Ports    []*TempPortCreate `json:"ports"`
		Metadata []*TempFields     `json:"metadata"`
		UserData *string           `json:"user_data,omitempty"`
		Disks    []*TempDisk       `json:"disks"`
		Floating *string           `json:"floating"`
	}{
		Name:     vm.Name,
		Cpu:      vm.Cpu,
		Ram:      vm.Ram,
		Vdc:      v.ID,
		Template: vm.Template.ID,
		Ports:    tempPorts,
		Metadata: tempFields,
		UserData: vm.UserData,
		Disks:    tempDisks,
		Floating: nil,
	}

	if vm.Floating != nil {
		args.Floating = vm.Floating.IpAddress
	}

	err := v.manager.Request("POST", "v1/vm", args, &vm)
	if err == nil {
		vm.manager = v.manager
		for idx := range vm.Ports {
			vm.Ports[idx].manager = v.manager
		}
		for idx := range vm.Disks {
			vm.Disks[idx].manager = v.manager
		}
		if vm.Floating != nil {
			vm.Floating.manager = v.manager
		}
	}

	return err
}

func (v *Vdc) CreateDisk(disk *Disk) error {
	args := &struct {
		Name           string  `json:"name"`
		Vdc            *string `json:"vdc,omitempty"`
		Vm             *string `json:"vm,omitempty"`
		Size           int     `json:"size"`
		StorageProfile string  `json:"storage_profile"`
	}{
		Name:           disk.Name,
		Vdc:            &v.ID,
		Vm:             nil,
		Size:           disk.Size,
		StorageProfile: disk.StorageProfile.ID,
	}

	if disk.Vm != nil {
		args.Vm = &disk.Vm.ID
		args.Vdc = nil
	}

	err := v.manager.Request("POST", "v1/disk", args, &disk)
	if err == nil {
		disk.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateEmptyPort(port *Port) (err error) {
	args := &struct {
		manager   *Manager
		ID        string  `json:"id"`
		IpAddress *string `json:"ip_address,omitempty"`
		Network   string  `json:"network"`
	}{
		ID:        port.ID,
		IpAddress: port.IpAddress,
		Network:   port.Network.ID,
	}

	err = v.manager.Request("POST", "v1/port", args, &port)
	return
}

func (v Vdc) WaitLock() (err error) {
	path, _ := url.JoinPath("v1/vdc", v.ID)
	return loopWaitLock(v.manager, path)
}
