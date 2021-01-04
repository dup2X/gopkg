package discovery

import (
	"strconv"
	"strings"

	"git.xiaojukeji.com/platform-ha/nodemgr-go"
)

//server status
const (
	HEALTHY   = nodemgr.HEALTHY
	UNHEALTHY = nodemgr.UNHEALTHY
)

var _ Balancer = new(nodeAgent)

type nodeAgent struct {
	serviceName string
}

//NodeOption ...
type NodeOption struct {
	WorkerCycle      int
	HealthyThreshold int64
	MaxCooldownTime  int64
	MinHealthyRatio  float64
}

//NewNodeMgr generate a nodemgr_type balancer object
func NewNodeMgr(ns string, hosts []string, nodeOpt *NodeOption) (Balancer, error) {
	service := make(map[string]nodemgr.ServiceConf)
	service[ns] = nodemgr.ServiceConf{
		Hosts:            hosts,
		HealthyThreshold: nodeOpt.HealthyThreshold,
		MaxCooldownTime:  nodeOpt.MaxCooldownTime,
		MinHealthyRatio:  nodeOpt.MinHealthyRatio,
	}
	err := nodemgr.InitV2(&nodemgr.ConfigJson{
		WorkerCycle: nodeOpt.WorkerCycle,
		Services:    service,
	})

	if err != nil {
		return nil, err
	}

	ag := &nodeAgent{}
	ag.Start(ns)
	return ag, nil
}

func (a *nodeAgent) Start(namespace string) error {
	a.serviceName = namespace
	return nil
}

func (a *nodeAgent) Up(addr string) (down func(error)) {
	return func(error) {}
}

func (a *nodeAgent) Get() (addr string, err error) {
	addr, err = nodemgr.GetNode(a.serviceName, "")
	if addr != "" {
		return addr, nil
	}

	return "", errNotFoundHostByNodemgr
}

func (a *nodeAgent) GetHostPort() (host string, port int, err error) {
	addr, err := a.Get()
	strArray := strings.Split(addr, ":")
	if len(strArray) < 2 {
		return "", 0, errNotFoundHostByNodemgr
	}

	port, _ = strconv.Atoi(strArray[1])
	return strArray[0], port, err
}

func (a *nodeAgent) Vote(addr string, voteType int) error {
	return nodemgr.Vote(a.serviceName, addr, voteType)
}

func (a *nodeAgent) Notify() <-chan []string { return nil }

func (a *nodeAgent) Close() error { return nil }
