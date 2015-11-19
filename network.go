package main

import (
	"errors"
	"fmt"
	"net"

	"gopkg.in/yaml.v2"
)

func ReadNetworks(data []byte) ([]*network, error) {
	var networks []*network

	if err := yaml.Unmarshal(data, &networks); err != nil {
		return networks, err
	}

	return networks, nil
}

type network struct {
	Name          string         `yaml:"name" json:"name"`
	Description   string         `yaml:"description" json:"description"`
	CIDR          string         `yaml:"cidr" json:"cidr"`
	DC            string         `yaml:"dc" json:"dc"`
	Gateway       net.IP         `yaml:"gateway" json:"gateway"`
	DNS           []net.IP       `yaml:"dns" json:"dns"`
	Vlan          vlan           `yaml:"vlan" json:"vlan"`
	DHCP          []rng          `yaml:"dhcp" json:"dhcp"`
	ForeignRanges []foreignRange `yaml:"foreign_ranges" json:"foreign_ranges"`
	Utilization   utilization    `yaml:"utilization" json:"utilization"`
}

type utilization struct {
	Total       int `yaml:"total" json:"total"`
	Free        int `yaml:"free" json:"free"`
	Used        int `yaml:"used" json:"used"`
	FreePercent int `yaml:"free_percent" json:"free_percent"`
	UsedPercent int `yaml:"used_percent" json:"used_percent"`
}

func (n network) Contains(ip net.IP) bool {
	_, ipnet, err := net.ParseCIDR(n.CIDR)
	if err != nil {
		return false
	}
	return ipnet.Contains(ip)
}
func (n network) Expand() ([]net.IP, error) {
	out := []net.IP{}
	ip, ipnet, err := net.ParseCIDR(n.CIDR)
	if err != nil {
		return out, err
	}

	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); ip = nextIP(ip) {
		out = append(out, ip)
	}

	return out, nil
}

func (n network) ExpandManaged() ([]net.IP, error) {
	out := []net.IP{}

	all, err := n.Expand()
	if err != nil {
		return out, err
	}

	// assume that last ip is broadcast an first is network
	l := len(all) - 1
	all = all[1:l]

	substract := []net.IP{}

	for _, dr := range n.DHCP {
		substract = append(substract, dr.Expand()...)
	}

	for _, fr := range n.ForeignRanges {
		substract = append(substract, fr.Rng.Expand()...)
	}

	for _, ip := range all {
		managed := true
		for _, sip := range substract {
			if ip.Equal(sip) {
				managed = false
				break
			}
		}
		if managed {
			out = append(out, ip)
		}
	}

	return out, nil
}

type rng struct {
	Start net.IP `yaml:"start" json:"start"`
	End   net.IP `yaml:"end" json:"end"`
}

func (r rng) Expand() []net.IP {
	ip := dupIP(r.Start)
	out := []net.IP{ip}

	for !ip.Equal(r.End) {
		ip = nextIP(ip)
		out = append(out, ip)
	}

	return out
}

func dupIP(ip net.IP) net.IP {
	// To save space, try and only use 4 bytes
	if x := ip.To4(); x != nil {
		ip = x
	}
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

func nextIP(ip net.IP) net.IP {
	next := dupIP(ip)
	for j := len(next) - 1; j >= 0; j-- {
		next[j]++
		if next[j] > 0 {
			break
		}
	}
	return next
}

type vlan struct {
	Name string `yaml:"name" json:"name"`
	Id   int64  `yaml:"id" json:"id"`
}

type foreignRange struct {
	Description string `yaml:"description" json:"description"`
	Rng         rng    `yaml:"range" json:"range"`
}

func tokenizeIP(ip net.IP) ([]uint, error) {
	p := ip

	if p4 := p.To4(); len(p4) == net.IPv4len {
		return []uint{uint(p4[0]),
				uint(p4[1]),
				uint(p4[2]),
				uint(p4[3])},
			nil
	}

	msg := fmt.Sprintf("%v is not an IPv4 address", ip)
	var tokens []uint
	return tokens, errors.New(msg)
}
