package main

import (
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

type Resolver struct {
	addrs  []*string
	mu     sync.Mutex
	OnRecv func([]*Response)
	OnIdle func()
	Debug  bool
}

func NewResolver() *Resolver {
	rand.Seed(time.Now().UnixNano())
	return &Resolver{
		addrs:  []*string{},
		OnRecv: nil,
		OnIdle: nil,
		Debug:  false,
	}
}

func (r *Resolver) AddAddr(addr string) error {
	r.mu.Lock()
	r.addrs = append(r.addrs, &addr)
	r.mu.Unlock()
	return nil
}

func (r *Resolver) RemoveAddr(addr string) error {
	r.mu.Lock()
	for i, a := range r.addrs {
		if a == &addr {
			r.addrs = r.addrs[:i+copy(r.addrs[i:], r.addrs[i+1:])]
		}
	}
	r.mu.Unlock()
	return nil
}

func (r *Resolver) Run() error {
	var wg sync.WaitGroup
	var err error

	for _, addr := range r.addrs {
		wg.Add(1)
		go func(addr *string, err error) {
			defer wg.Done()
			re, err := Resolv(*addr)
			r.OnRecv(re)
		}(addr, err)
	}
	wg.Wait()
	r.OnIdle()
	return err
}

type Response struct {
	A     string `json:"aptr"`
	Addr  net.IP `json:"address"`
	TXT   string `json:"txt"`
	CNAME string `json:"cname"`
}

func Resolv(in string) ([]*Response, error) {
	ip := net.ParseIP(in)
	if ip != nil {
		// Input is an IP Address
		return resolvIP(ip)
	} else {
		// Input is a CNAME or an APTR
		return resolvName(in)
	}
}

func resolvIP(ip net.IP) ([]*Response, error) {
	var sets []*Response
	aptrs, err := net.LookupAddr(ip.String())
	if err != nil {
		return sets, err
	}
	for _, aptr := range aptrs {
		txt := resolvTXT(aptr)
		sets = append(sets, &Response{
			Addr: ip,
			A:    aptr,
			TXT:  txt,
		})
	}
	return sets, nil
}

func resolvTXT(addr string) string {
	txts, _ := net.LookupTXT(addr)
	return strings.Join(txts, ", ")
}

func resolvName(name string) ([]*Response, error) {
	var sets []*Response
	isCNAME := true

	// append a dot if there is none
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}

	// check if there name is a CNAME
	aptr, err := net.LookupCNAME(name)
	if err != nil {
		return sets, err
	}
	if aptr == name {
		isCNAME = false
	}

	// get IP address
	addrs, err := net.LookupHost(name)
	if err != nil {
		return sets, err
	}

	// return a DNSSet for each address
	for _, addr := range addrs {
		txt := resolvTXT(name)
		if isCNAME {
			sets = append(sets, &Response{
				Addr:  net.ParseIP(addr),
				A:     aptr,
				CNAME: name,
				TXT:   txt,
			})

		} else {
			sets = append(sets, &Response{
				Addr: net.ParseIP(addr),
				A:    name,
				TXT:  txt,
			})
		}
	}
	return sets, nil
}
