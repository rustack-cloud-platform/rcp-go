package rustack

import (
	"fmt"
)

type RouterFirewallRule struct {
	manager         *Manager
	routerId        string
	ID              string `json:"id"`
	Name            string `json:"name"`
	Direction       string `json:"direction"`
	DestinationIp   string `json:"destination_ip"`
	DstPortRangeMax *int   `json:"dst_port_range_max"`
	DstPortRangeMin *int   `json:"dst_port_range_min"`
	SourceIp        string `json:"source_ip"`
	SrcPortRangeMax *int   `json:"src_port_range_max"`
	SrcPortRangeMin *int   `json:"src_port_range_min"`
	Protocol        string `json:"protocol"`
	Locked          bool   `json:"locked"`
}

func NewRouterFirewallRule(
	name string,
	protocol string,
	direction string,
	destinationIp string,
	dstPortRangeMax, dstPortRangeMin int,
	sourceIp string,
	srcPortRangeMax, srcPortRangeMin int,
) RouterFirewallRule {
	d := RouterFirewallRule{
		Name:            name,
		DestinationIp:   destinationIp,
		Direction:       direction,
		DstPortRangeMax: &dstPortRangeMax,
		DstPortRangeMin: &dstPortRangeMin,
		SourceIp:        sourceIp,
		SrcPortRangeMax: &srcPortRangeMax,
		SrcPortRangeMin: &srcPortRangeMin,
		Protocol:        protocol,
	}
	return d
}

func (r *Router) CreateFirewallRule(firewallRule *RouterFirewallRule) (err error) {
	args := &struct {
		manager         *Manager
		ID              string `json:"id"`
		Name            string `json:"name"`
		Protocol        string `json:"protocol"`
		Direction       string `json:"direction"`
		DestinationIp   string `json:"destination_ip"`
		DstPortRangeMax *int   `json:"dst_port_range_max"`
		DstPortRangeMin *int   `json:"dst_port_range_min"`
		SourceIp        string `json:"source_ip"`
		SrcPortRangeMax *int   `json:"src_port_range_max"`
		SrcPortRangeMin *int   `json:"src_port_range_min"`
	}{
		ID:              firewallRule.ID,
		Name:            firewallRule.Name,
		Protocol:        firewallRule.Protocol,
		Direction:       firewallRule.Direction,
		DestinationIp:   firewallRule.DestinationIp,
		DstPortRangeMax: nil,
		DstPortRangeMin: nil,
		SourceIp:        firewallRule.SourceIp,
		SrcPortRangeMax: nil,
		SrcPortRangeMin: nil,
	}

	if firewallRule.Protocol == "tcp" || firewallRule.Protocol == "udp" {
		args.DstPortRangeMax = firewallRule.DstPortRangeMax
		args.DstPortRangeMin = firewallRule.DstPortRangeMin
		args.SrcPortRangeMax = firewallRule.SrcPortRangeMax
		args.SrcPortRangeMin = firewallRule.SrcPortRangeMin
	}

	path := fmt.Sprintf("v1/router/%s/firewall_rule", r.ID)
	err = r.manager.Request("POST", path, args, &firewallRule)
	if err != nil {
		return err
	}
	firewallRule.manager = r.manager
	firewallRule.routerId = r.ID
	return
}

func (r *Router) GetFirewallRuleById(firewallRuleId string) (firewallRule *RouterFirewallRule, err error) {
	path := fmt.Sprintf("v1/router/%s/firewall_rule/%s", r.ID, firewallRuleId)
	err = r.manager.Get(path, Defaults(), &firewallRule)
	if err != nil {
		return
	}
	firewallRule.manager = r.manager
	firewallRule.routerId = r.ID
	return
}

func (r *Router) GetFirewallRules() (firewallRules []*RouterFirewallRule, err error) {
	path := fmt.Sprintf("v1/router/%s/firewall_rule", r.ID)
	err = r.manager.Get(path, Defaults(), &firewallRules)
	if err != nil {
		return
	}
	return
}

func (f *RouterFirewallRule) Update() (err error) {
	path := fmt.Sprintf("v1/router/%s/firewall_rule/%s", f.routerId, f.ID)
	return f.manager.Request("PUT", path, f, &f)
}

func (f *RouterFirewallRule) Delete() (err error) {
	path := fmt.Sprintf("v1/router/%s/firewall_rule/%s", f.routerId, f.ID)
	return f.manager.Delete(path, Defaults(), nil)
}

func (f RouterFirewallRule) WaitLock() (err error) {
	path := fmt.Sprintf("v1/router/%s/firewall_rule/%s", f.routerId, f.ID)
	return loopWaitLock(f.manager, path)
}
