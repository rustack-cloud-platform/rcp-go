package rustack

import (
	"fmt"
)

type Router struct {
	manager   *Manager
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Vdc       struct {
		Id string `json:"id"`
	} `json:"vdc"`
	Ports    []*Port   `json:"ports"`
	Floating *Floating `json:"floating"`
	Locked   bool      `json:"locked"`
}

func NewRouter(name string) Router {
	r := Router{Name: name}
	return r
}

func (m *Manager) GetRouters(extraArgs ...Arguments) (routers []*Router, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/router"
	err = m.GetItems(path, args, &routers)
	for i := range routers {
		routers[i].manager = m
		for x := range routers[i].Ports {
			routers[i].Ports[x].manager = m
		}
	}
	return
}

func (v *Vdc) GetRouters(extraArgs ...Arguments) (routers []*Router, err error) {
	args := Arguments{
		"vdc": v.ID,
	}
	args.merge(extraArgs)
	routers, err = v.manager.GetRouters(args)
	return
}

func (m *Manager) GetRouter(id string) (router *Router, err error) {
	path := fmt.Sprintf("v1/router/%s", id)
	err = m.Get(path, Defaults(), &router)
	if err != nil {
		return
	}
	router.manager = m
	return
}

func (r Router) WaitLock() (err error) {
	path := fmt.Sprintf("v1/router/%s", r.ID)
	return loopWaitLock(r.manager, path)
}

func (r *Router) AddPort(port *Port) error {
	type TempPortCreate struct {
		Router      string   `json:"router"`
		Network     string   `json:"network"`
		IpAddress   *string  `json:"ip_address,omitempty"`
		FwTemplates []string `json:"fw_templates"`
	}

	var fwTemplates = make([]string, len(port.FirewallTemplates))
	for i, fwTemplate := range port.FirewallTemplates {
		fwTemplates[i] = fwTemplate.ID
	}
	args := &TempPortCreate{
		Router:      r.ID,
		Network:     port.Network.ID,
		IpAddress:   port.IpAddress,
		FwTemplates: fwTemplates,
	}

	err := r.manager.Post("v1/port", args, &port)
	if err == nil {
		port.manager = r.manager
	}

	return err
}

func (r *Router) Delete() error {
	path := fmt.Sprintf("v1/router/%s", r.ID)
	return r.manager.Delete(path, Defaults(), r)
}

func (r *Router) Update() error {
	args := &struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsDefault bool   `json:"is_default"`
		Vdc       struct {
			Id string `json:"id"`
		} `json:"vdc"`
		Ports    []*Port `json:"ports"`
		Floating *string `json:"floating"`
	}{
		ID:        r.ID,
		Name:      r.Name,
		IsDefault: r.IsDefault,
		Vdc:       r.Vdc,
		Ports:     r.Ports,
	}
	if r.Floating == nil {
		args.Floating = nil
	} else {
		args.Floating = &r.Floating.ID
	}
	path := fmt.Sprintf("v1/router/%s", r.ID)
	r.WaitLock()
	return r.manager.Put(path, args, r)
}
