package ratelimit

import (
	"time"
)

const (
	defaultFillInterval = time.Millisecond * 10
)

//Bucket :Bucket
type Bucket struct {
	fillInterval time.Duration
	rate         int64
	quantum      int64
	ch           chan struct{}
}

//New :new
func New(rate int64) *Bucket {
	return NewWithFillInterval(defaultFillInterval, rate)
}

//NewWithFillInterval :new with fill interval
func NewWithFillInterval(fillInterval time.Duration, rate int64) *Bucket {
	if fillInterval > time.Second {
		panic("fill_duration shouldn't be more than second")
	}
	quantum := int64(time.Second / fillInterval)
	bt := &Bucket{
		fillInterval: fillInterval,
		rate:         rate,
		quantum:      quantum,
		ch:           make(chan struct{}, rate),
	}
	for i := int64(0); i < bt.quantum; i++ {
		bt.ch <- struct{}{}
	}
	go bt.run()
	return bt

}

func (bt *Bucket) run() {
	tk := time.NewTicker(bt.fillInterval)
	for range tk.C {
		for i := int64(0); i < bt.quantum; i++ {
			select {
			case bt.ch <- struct{}{}:
			default:
			}
		}
	}
}

// GetTicket :get the ticket
func (bt *Bucket) GetTicket() bool {
	select {
	case <-bt.ch:
		return true
	default:
		return false
	}
}
