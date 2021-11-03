package rustack

import "fmt"

type Vdc struct {
	manager    *Manager
	ID         string     `json:"id"`
	Name       string     `json:"name"`
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
	path := fmt.Sprintf("v1/vdc/%s", id)
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

	err := p.manager.Post("v1/vdc", args, &vdc)
	if err == nil {
		vdc.manager = p.manager
	}

	return err
}

func (v *Vdc) Rename(name string) error {
	path := fmt.Sprintf("v1/vdc/%s", v.ID)
	return v.manager.Put(path, Arguments{"name": name}, v)
}

func (v *Vdc) Delete() error {
	path := fmt.Sprintf("v1/vdc/%s", v.ID)
	return v.manager.Delete(path, Defaults(), v)
}

func (v *Vdc) CreateNetwork(network *Network) error {
	args := Arguments{
		"name": network.Name,
		"vdc":  v.ID,
	}

	err := v.manager.Post("v1/network", args, &network)
	if err == nil {
		network.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateRouter(router *Router, ports ...*Port) error {
	type TempPortCreate struct {
		Network   string  `json:"network"`
		IpAddress *string `json:"ip_address,omitempty"`
	}

	tempPorts := make([]*TempPortCreate, len(ports))
	for idx := range ports {
		tempPorts[idx] = &TempPortCreate{Network: ports[idx].Network.ID, IpAddress: ports[idx].IpAddress}
	}

	args := &struct {
		Name     string            `json:"name"`
		Vdc      string            `json:"vdc"`
		Ports    []*TempPortCreate `json:"ports"`
		Floating string            `json:"floating"`
	}{
		Name:     router.Name,
		Vdc:      v.ID,
		Ports:    tempPorts,
		Floating: "RANDOM_FIP",
	}

	err := v.manager.Post("v1/router", args, &router)
	if err == nil {
		router.manager = v.manager
	}

	return err
}

func (v *Vdc) CreateVm(vm *Vm) error {
	type TempPortCreate struct {
		Network     string    `json:"network"`
		IpAddress   *string   `json:"ip_address,omitempty"`
		FwTemplates []*string `json:"fw_templates"`
	}

	tempPorts := make([]*TempPortCreate, len(vm.Ports))
	for idx := range vm.Ports {
		var fwTemplates = make([]*string, 0)
		for _, fwTemplate := range vm.Ports[idx].FirewallTemplates {
			fwTemplates = append(fwTemplates, &fwTemplate.ID)
		}
		tempPorts[idx] = &TempPortCreate{Network: vm.Ports[idx].Network.ID, IpAddress: vm.Ports[idx].IpAddress, FwTemplates: fwTemplates}
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

	err := v.manager.Post("v1/vm", args, &vm)
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

	err := v.manager.Post("v1/disk", args, &disk)
	if err == nil {
		disk.manager = v.manager
	}

	return err
}
