package utils

import "sync"

type WaitGroupMap struct {
	mu sync.RWMutex
	m  map[string]*sync.WaitGroup
}

func NewWaitGroups() *WaitGroupMap {
	return &WaitGroupMap{
		m: make(map[string]*sync.WaitGroup),
	}
}

func (w *WaitGroupMap) Add(delta int, key string) {
	wg := w.getOrAddWaitGroup(key)
	wg.Add(delta)
}

func (w *WaitGroupMap) AddMany(delta int, keys ...string) {
	switch len(keys) {
	case 0:
		return
	case 1:
		w.Add(delta, keys[0])
		return
	case 2:
		w.Add(delta, keys[0])
		w.Add(delta, keys[1])
		return
	}

	for _, key := range keys {
		wg := w.getOrAddWaitGroup(key)
		wg.Add(delta)
	}
}

func (w *WaitGroupMap) Done(key string) {
	w.mu.RLock()
	wg := w.m[key]
	w.mu.RUnlock()
	if wg != nil {
		wg.Done()
	}
}

func (w *WaitGroupMap) DoneMany(keys ...string) { //nolint:dupl
	switch len(keys) {
	case 0:
		return
	case 1:
		w.Done(keys[0])
		return
	case 2:
		w.Done(keys[0])
		w.Done(keys[1])
		return
	}

	wgs := make([]*sync.WaitGroup, 0, len(keys))

	w.mu.RLock()
	for _, key := range keys {
		wg := w.m[key]
		if wg != nil {
			wgs = append(wgs, wg)
		}
	}
	w.mu.RUnlock()

	for _, wg := range wgs {
		wg.Done()
	}
}

func (w *WaitGroupMap) Wait(key string) {
	w.mu.RLock()
	wg := w.m[key]
	w.mu.RUnlock()
	if wg != nil {
		wg.Wait()
	}
}

func (w *WaitGroupMap) WaitMany(keys ...string) { //nolint:dupl
	switch len(keys) {
	case 0:
		return
	case 1:
		w.Wait(keys[0])
		return
	case 2:
		w.Wait(keys[0])
		w.Wait(keys[1])
		return
	}

	wgs := make([]*sync.WaitGroup, 0, len(keys))

	w.mu.RLock()
	for _, key := range keys {
		wg := w.m[key]
		if wg != nil {
			wgs = append(wgs, wg)
		}
	}
	w.mu.RUnlock()

	for _, wg := range wgs {
		wg.Wait()
	}
}

func (w *WaitGroupMap) getOrAddWaitGroup(key string) *sync.WaitGroup {
	w.mu.RLock()
	wg := w.m[key]
	w.mu.RUnlock()

	if wg == nil {
		w.mu.Lock()
		wg = w.m[key]
		if wg == nil {
			wg = &sync.WaitGroup{}
			w.m[key] = wg
		}
		w.mu.Unlock()
	}

	return wg
}
