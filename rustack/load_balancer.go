package rustack

import "fmt"

type LoadBalancer struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Locked  bool   `json:"locked"`
	Vdc     *Vdc   `json:"vdc"`
	JobId   string `json:"job_id"`

	Kubernetes *Kubernetes `json:"kubernetes"`
	Port       *Port       `json:"port"`
	Floating   *Port       `json:"floating"`
	AutoIp     bool        `json:"autoIp"`
}

type LoadBalancerPool struct {
	manager *Manager
	ID      string `json:"id"`
	Locked  bool   `json:"locked"`

	Port               int           `json:"port"`
	Connlimit          int           `json:"connlimit"`
	Members            []*PoolMember `json:"members"`
	Method             string        `json:"method"`
	Protocol           string        `json:"protocol"`
	SessionPersistence *string       `json:"session_persistence"`
}

type PoolMember struct {
	ID     string `json:"id"`
	Port   int    `json:"port"`
	Weight int    `json:"weight"`
	Vm     *Vm    `json:"vm"`
}

func (m *Manager) GetLoadBalancers(extraArgs ...Arguments) (lbaasList []*LoadBalancer, err error) {
	args := Defaults()
	args.merge(extraArgs)

	path := "v1/lbaas"
	err = m.GetItems(path, args, &lbaasList)
	for i := range lbaasList {
		lbaasList[i].manager = m
		lbaasList[i].Port.manager = m
		lbaasList[i].Vdc.manager = m
		if lbaasList[i].Floating != nil {
			lbaasList[i].Floating.manager = m
		}
	}
	return
}

func (m *Manager) GetLoadBalancer(id string) (lbaas LoadBalancer, err error) {
	path := fmt.Sprintf("v1/lbaas/%s", id)
	err = m.Get(path, Defaults(), &lbaas)
	if err != nil {
		return
	}
	return
}

func NewLoadBalancer(name string, vdc *Vdc, port *Port, floating *Port, autoIp bool) *LoadBalancer {
	return &LoadBalancer{
		manager:  vdc.manager,
		Name:     name,
		Vdc:      vdc,
		Port:     port,
		Floating: floating,
		AutoIp:   autoIp,
	}
}

func (lb *LoadBalancer) Create() (err error) {
	type customPort struct {
		ID                string     `json:"id"`
		IpAddress         *string    `json:"ip_address,omitempty"`
		Network           string     `json:"network"`
		FirewallTemplates *string    `json:"fw_templates,omitempty"`
		Connected         *Connected `json:"connected"`
	}
	lbCreate := &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Vdc  string `json:"vdc"`

		Kubernetes *Kubernetes `json:"kubernetes"`
		Port       customPort  `json:"port"`
		Floating   string      `json:"floating"`
		AutoIp     bool        `json:"autoIp"`
	}{
		Name: lb.Name,
		Vdc:  lb.Vdc.ID,
		Port: customPort{
			ID:                lb.Port.ID,
			IpAddress:         lb.Port.IpAddress,
			Network:           lb.Port.Network.ID,
			FirewallTemplates: nil,
			Connected:         lb.Port.Connected,
		},
		Kubernetes: lb.Kubernetes,
		Floating:   lb.Floating.ID,
		AutoIp:     lb.AutoIp,
	}
	err = lb.manager.Post("v1/lbaas", lbCreate, &lb)
	return
}

func (lb *LoadBalancer) Delete() (err error) {
	path := fmt.Sprintf("v1/lbaas/%s", lb.ID)
	err = lb.manager.Delete(path, Defaults(), &lb)
	return
}

func (lb *LoadBalancer) DeletePools() (err error) {

	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
	var pools []*LoadBalancerPool
	err = lb.manager.GetSubItems(path, Arguments{}, &pools)
	if err != nil {
		return err
	}
	for _, pool := range pools {
		path = fmt.Sprintf("v1/lbaas/%s/pool/%s", lb.ID, pool.ID)
		err = lb.manager.Delete(path, Defaults(), Defaults())
		if err != nil {
			return err
		}
	}
	return
}

func (lb *LoadBalancer) CreatePool(pool *LoadBalancerPool) (err error) {

	type poolMember struct {
		Port   int    `json:"port"`
		Weight int    `json:"weight"`
		Vm     string `json:"vm"`
	}
	type createPool struct {
		Port               int           `json:"port"`
		Connlimit          int           `json:"connlimit"`
		Members            []*poolMember `json:"members"`
		Method             string        `json:"method"`
		Protocol           string        `json:"protocol"`
		SessionPersistence *string       `json:"session_persistence"`
	}

	var members []*poolMember
	for _, member := range pool.Members {
		members = append(members, &poolMember{
			Port:   member.Port,
			Weight: member.Weight,
			Vm:     member.Vm.ID,
		})
	}

	lbCreatePool := createPool{
		Port:               pool.Port,
		Connlimit:          pool.Connlimit,
		Members:            members,
		Method:             pool.Method,
		Protocol:           pool.Protocol,
		SessionPersistence: pool.SessionPersistence,
	}
	path := fmt.Sprintf("v1/lbaas/%s/pool", lb.ID)
	err = lb.manager.Post(path, lbCreatePool, &pool)
	return
}

func (lb LoadBalancer) WaitLock() (err error) {
	path := fmt.Sprintf("v1/lbaas/%s", lb.ID)
	return loopWaitLock(lb.manager, path)
}
