package main

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/tatsushid/go-fastping"
)

type ResultSet struct {
	IP           net.IP `json:"ip"`
	Name         string `json:"name"`
	Desc         string `json:"desc"`
	Pingable     bool   `json:"pingable"`
	Lock         Lock   `json:"lock"`
	Free         bool   `json:"free"`
	ForeignRange string `json:"foreign_range"`
	Unmanaged    string `json:"unmanaged"`
}

func (rs ResultSet) Used() bool {
	return rs.Pingable || rs.Name != "" || rs.Lock.Locked()
}

type check struct {
	sync.RWMutex
	results     map[string]*ResultSet
	utilization utilization
}

func NewCheck(ips detailedIP) *check {
	var c check
	c.results = make(map[string]*ResultSet)
	c.utilization = utilization{}
	for ip, details := range ips {
		res := ResultSet{
			IP:           details.IP,
			Name:         "",
			Desc:         "",
			Pingable:     false,
			Lock:         Lock{},
			Free:         false,
			ForeignRange: "",
			Unmanaged:    details.Unmanaged,
		}
		c.results[ip] = &res

	}
	return &c
}

func (c *check) isResolvable() {
	r := NewResolver()
	for ip, _ := range c.results {
		r.AddAddr(ip)
	}
	r.OnRecv = func(resp []*Response) {
		if len(resp) > 0 {
			c.Lock()
			c.results[resp[0].Addr.String()].Name = resp[0].A
			c.results[resp[0].Addr.String()].Desc = resp[0].TXT
			c.Unlock()
		}
	}
	r.OnIdle = func() {}
	err := r.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func (c *check) isPingable() {
	p := fastping.NewPinger()
	for ip, _ := range c.results {
		p.AddIP(ip)
	}
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		c.Lock()
		c.results[addr.String()].Pingable = true
		c.Unlock()
	}
	p.OnIdle = func() {}
	err := p.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func (c *check) isLocked() {
	locker.Clean()
	for ip, r := range c.results {
		c.Lock()
		r.Lock = locker.Get(ip)
		c.Unlock()
	}
}

func (c *check) isForeign() {
	//	for ip, r := range c.results {
	//		c.Lock()
	//
	//		c.Unlock()
	//	}
}

func (c *check) getFree() {
	defer c.Unlock()
	c.Lock()
	total := len(c.results)
	free := 0
	used := 0
	for _, r := range c.results {
		if r.Used() {
			used += 1
			r.Free = false
		} else {
			free += 1
			r.Free = true
		}
	}

	c.utilization.Total = total
	c.utilization.Used = used
	c.utilization.Free = free
	c.utilization.UsedPercent = used * 100 / total
	c.utilization.FreePercent = free * 100 / total
}

func (c *check) Run() {
	c.isResolvable()
	c.isPingable()
	c.isLocked()
	c.isForeign()
	c.getFree()
}
