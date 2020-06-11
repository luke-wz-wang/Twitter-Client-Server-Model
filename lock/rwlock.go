// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import "sync"

type RWLock struct {
	m *sync.Mutex
	w bool
	readerCount int64
	c *sync.Cond
}

func NewRWLock() *RWLock{
	m := sync.Mutex{}
	w := false
	var count int64
	c := sync.NewCond(&m)
	return &RWLock{&m, w, count, c}
}

// writer locks
// locks for writer
func (lock *RWLock) Lock(){
	lock.m.Lock()
	// check for active writer/readers; wait if true
	for lock.w || lock.readerCount > 0{
		lock.c.Wait()
	}
	lock.w = true
	lock.m.Unlock()
}
// unlocks for writer
func (lock *RWLock) Unlock(){
	lock.m.Lock()
	lock.w = false
	lock.c.Broadcast()
	lock.m.Unlock()
}

// reader locks
// locks for readers
func (lock *RWLock) RLock(){
	lock.m.Lock()
	// check for active writer or whether max limit of readers has been met; wait if true
	for lock.w || lock.readerCount >= 32{
		lock.c.Wait()
	}
	lock.readerCount++
	lock.m.Unlock()
}

// unlocks for readers
// broadcast if there is no reader anymore
func (lock *RWLock) RUnlock(){
	lock.m.Lock()
	lock.readerCount--
	if lock.readerCount == 0{
		lock.c.Broadcast()
	}
	lock.m.Unlock()
}