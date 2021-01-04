//Package discovery defined balancer
// we use interface ...
package discovery

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dup2X/gopkg/config"
)

const (
	splitChar = ","
)

//balancer type
const (
	LOCALTYPE = iota
	DISFTYPE
	NODEMGRTYPE
)

//Balancer represents load balancing for downstream
type Balancer interface {
	Start(namespace string) error
	Up(addr string) (down func(error))
	Get() (addr string, err error)
	GetHostPort() (host string, port int, err error)
	Vote(addr string, voteType int) error
	Notify() <-chan []string
	Close() error
}

var _ Balancer = new(serverList)

//NewBalancer generate a balancer object by type
func NewBalancer(blancerType int, ns string, hosts []string, nodeOpt *NodeOption) (Balancer, error) {
	switch blancerType {
	case LOCALTYPE:
		return NewWithHosts(hosts), nil
	case NODEMGRTYPE:
		return NewNodeMgr(ns, hosts, nodeOpt)
	default:
		return nil, errBalancerType
	}
}

//New generate a local_type balancer object by Sectioner
func New(sec config.Sectioner) (Balancer, error) {
	hs, err := sec.GetString("hosts")
	if err != nil {
		return nil, err
	}
	hosts := strings.Split(hs, splitChar)
	br := NewWithHosts(hosts)
	return br, nil
}

//NewWithHosts generate a local_type balancer object by hosts
func NewWithHosts(hosts []string) Balancer {
	br := newServerList(hosts)
	br.Start("local")
	return br
}

// TODO sd
type serverList struct {
	namespace string
	waitCh    chan []string

	mu    *sync.Mutex
	next  int
	addrs []string
	state map[string]bool
}

var (
	errEmptyAddrs            = fmt.Errorf("there has no addr to pick")
	errFullNotifyChan        = fmt.Errorf("notify chan is full")
	errNotFoundHostByDisf    = fmt.Errorf("disf not return any host")
	errNotFoundHostByNodemgr = fmt.Errorf("nodemgr not return any host")
	errBalancerType          = fmt.Errorf("invalid balacer type")
)

func newServerList(addrs []string) *serverList {
	return &serverList{
		waitCh: make(chan []string, 8),
		addrs:  addrs,
		mu:     new(sync.Mutex),
		state:  make(map[string]bool),
	}
}

func (sl *serverList) Start(namespace string) error {
	for _, addr := range sl.addrs {
		sl.state[addr] = true
	}
	sl.namespace = namespace
	return nil
}

func (sl *serverList) Up(addr string) (down func(error)) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if _, ok := sl.state[addr]; !ok {
		sl.addrs = append(sl.addrs, addr)
	}
	sl.state[addr] = true
	down = func(err error) {
		return
	}
	return
}

func (sl *serverList) Get() (addr string, err error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	if len(sl.addrs) == 0 {
		return "", errEmptyAddrs
	}
	if sl.next >= len(sl.addrs) {
		sl.next = 0
	}
	next := sl.next
	for {
		addr = sl.addrs[next]
		next = (next + 1) % len(sl.addrs)
		if st, ok := sl.state[addr]; st && ok {
			sl.next = next
			return
		}
		if next == sl.next {
			break
		}
	}
	addr = sl.addrs[next]
	return
}

func (sl *serverList) GetHostPort() (addr string, port int, err error) {
	addrs, err := sl.Get()
	strArray := strings.Split(addrs, ":")
	if len(strArray) < 2 {
		return "", 0, err
	}
	port, _ = strconv.Atoi(strArray[1])
	return strArray[0], port, err
}

func (sl *serverList) Vote(addr string, voteType int) error {
	return nil
}

func (sl *serverList) Notify() <-chan []string {
	return sl.waitCh
}

func (sl *serverList) Close() error {
	return nil
}

func (sl *serverList) set(addrs []string) error {
	select {
	case sl.waitCh <- addrs:
	default:
		return errFullNotifyChan
	}
	return nil
}
