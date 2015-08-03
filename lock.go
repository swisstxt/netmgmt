package main

import (
	"sync"
	"time"
)

type Lock struct {
	Comment     string    `json:"comment"`
	Owner       string    `json:"owner"`
	LockedUntil time.Time `json:"locked_until"`
}

func (l *Lock) Locked() bool {
	return l.Comment != "" || l.Owner != ""
}

type Locker struct {
	sync.RWMutex
	locks map[string]Lock
	ver   int64
	dur   int
}

func (l *Locker) Init(duration int) {
	l.dur = duration
	l.ver = 0
	l.locks = make(map[string]Lock)
}

func (l *Locker) Add(ip string, comment string, owner string) bool {
	l.Lock()
	defer l.Unlock()

	if _, ok := l.locks[ip]; !ok {
		lock := Lock{
			Comment:     comment,
			Owner:       owner,
			LockedUntil: time.Now().Add(time.Duration(l.dur) * time.Minute),
		}

		l.locks[ip] = lock
		l.ver = +1
		return true
	}
	return false
}

func (l *Locker) Delete(ip string) {
	l.Lock()
	defer l.Unlock()

	delete(l.locks, ip)
}

func (l *Locker) Get(ip string) Lock {
	l.RLock()
	defer l.RUnlock()

	return l.locks[ip]
}

func (l *Locker) Clean() {
	l.Lock()
	defer l.Unlock()

	for ip, lock := range l.locks {
		if lock.LockedUntil.Before(time.Now()) {
			delete(l.locks, ip)
		}
	}
}
