package rand

import (
	"math/rand"
	"sync"
	"time"
)

type lockedSource struct {
	sync.Mutex
	src rand.Source
}

//New :New rand
func New(seed int64) *rand.Rand {
	return rand.New(&lockedSource{src: rand.NewSource(seed)})
}

//NewSeeded :New Seeded
func NewSeeded() *rand.Rand {
	return New(time.Now().UnixNano())
}

// Int63 :
func (r *lockedSource) Int63() (n int64) {
	r.Lock()
	n = r.src.Int63()
	r.Unlock()
	return
}

// Seed :
func (r *lockedSource) Seed(seed int64) {
	r.Lock()
	r.src.Seed(seed)
	r.Unlock()
}
