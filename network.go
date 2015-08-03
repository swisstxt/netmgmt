package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
)

func ReadNetworks(data []byte) (*[]network, error) {
	var networks []network

	if err := json.Unmarshal(data, &networks); err != nil {
		return &networks, err
	}

	return &networks, nil
}

type network struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	CIDR          string         `json:"cidr"`
	DC            string         `json:"dc"`
	Gateway       net.IP         `json:"gateway"`
	DNS           []net.IP       `json:"dns"`
	Vlan          vlan           `json:"vlan"`
	DHCP          []rng          `json:"dhcp"`
	ForeignRanges []foreignRange `json:"foreign_ranges"`
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
	Start net.IP `json:"start"`
	End   net.IP `json:"end"`
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
	Name string `json:"name"`
	Id   int64  `json:"id"`
}

type foreignRange struct {
	Description string `json:"description"`
	Rng         rng    `json:"range"`
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
