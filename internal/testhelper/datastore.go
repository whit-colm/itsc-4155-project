package testhelper

import "sync"

type Datastore[K comparable, V any] struct {
	m sync.Map
}
