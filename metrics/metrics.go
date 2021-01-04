//Package metrics ...
package metrics

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dup2X/gopkg/config"

	gmetrics "github.com/rcrowley/go-metrics"
)

const (
	defaultSize          = 1024 * 4
	defaultReservoirSize = 1024 * 10

	defaultTTL = time.Second * 1
)

var (
	defaultClient *metricsStruct
	pool          sync.Pool
	reportTTL     int64
)

func init() {
	pool = sync.Pool{
		New: func() interface{} {
			return &packet{}
		},
	}
}

type transport interface {
	send(service string, snapshot map[string]interface{})
}

type command uint8

const (
	commandAdd command = iota
	commandAddErr
	commandDecr
	commandDecrErr
	commandElapsed
	commandGauge
	commandRPC
)

type packet struct {
	cmd     command
	key     string
	code    string
	c       int64
	elapsed int64
	fv      float64
	caller  string
	callee  string
}

type metricsStruct struct {
	Endpoint string
	Service  string
	Prefix   string

	durationUnit  time.Duration
	reportTTL     time.Duration
	reportTimeout time.Duration

	in  chan *packet
	r   gmetrics.Registry
	trs []transport

	debug bool
}

// NewDefault ...
func NewDefault() {
	defaultClient = &metricsStruct{
		Service:      "default",
		in:           make(chan *packet, defaultSize),
		r:            gmetrics.NewRegistry(),
		durationUnit: time.Millisecond * 1,
		reportTTL:    defaultTTL,
		trs:          []transport{newPrintClient()},
	}
	reportTTL = int64(defaultTTL) / 1e9
	go defaultClient.run()
}

// NewWithConfig ...
func NewWithConfig(sec config.Sectioner) error {
	service := sec.GetStringMust("service", "default")
	rep := sec.GetStringMust("report_duration", "1s")
	reportDur, err := time.ParseDuration(rep)
	if err != nil {
		return err
	}

	rt := sec.GetStringMust("report_timeout", "500ms")
	reportTimeout, err := time.ParseDuration(rt)
	if err != nil {
		return err
	}

	unit := sec.GetStringMust("latency_uint", "1ms")
	latencyUnit, err := time.ParseDuration(unit)
	if err != nil {
		return err
	}
	debug := sec.GetBoolMust("debug", false)

	trans := sec.GetStringMust("transport", "odin")
	var transports []transport
	for _, tr := range strings.Split(trans, ",") {
		switch tr {
		case "odin":
			odinURL, err := sec.GetString("odin_url")
			if err != nil {
				return err
			}
			transports = append(transports, newOdinClient(odinURL, reportTimeout))
		case "log":
			transports = append(transports, newPrintClient())
		}
	}

	defaultClient = &metricsStruct{
		Service:       service,
		reportTTL:     reportDur,
		durationUnit:  latencyUnit,
		reportTimeout: reportTimeout,
		trs:           transports,
		in:            make(chan *packet, defaultSize),
		r:             gmetrics.NewRegistry(),
		debug:         debug,
	}
	reportTTL = int64(reportDur) / 1e9
	go defaultClient.run()
	return nil
}

func (m *metricsStruct) run() {
	tk := time.NewTicker(m.reportTTL)
	for {
		select {
		case p := <-m.in:
			m.do(p)
			pool.Put(p)

		case <-tk.C:
			m.dump()
		}
	}
}

func (m *metricsStruct) dump() {
	snapshot := make(map[string]interface{})
	m.r.Each(func(key string, reg interface{}) {
		switch regInst := reg.(type) {
		case gmetrics.Counter:
			if regInst.Count() > 0 {
				snapshot[key] = regInst.Snapshot()
			}
			regInst.Clear()
		case gmetrics.Histogram:
			if regInst.Count() > 0 {
				snapshot[key] = regInst.Snapshot()
			}
			regInst.Clear()
		case gmetrics.GaugeFloat64:
			if regInst.Value() > 0 {
				snapshot[key] = regInst.Snapshot()
			}
		default:
			println("default")
		}
	})
	if len(snapshot) == 0 {
		return
	}
	for i := range m.trs {
		m.trs[i].send(m.Service, snapshot)
	}
}

func (m *metricsStruct) do(p *packet) {
	switch p.cmd {
	case commandAdd:
		cnt := gmetrics.NewCounter()
		if c, ok := m.r.GetOrRegister(p.key, cnt).(gmetrics.Counter); ok {
			c.Inc(p.c)
		}
	case commandAddErr:
		cnt := gmetrics.NewCounter()
		if c, ok := m.r.GetOrRegister(genErrKey(p.key, p.code), cnt).(gmetrics.Counter); ok {
			c.Inc(p.c)
		}
	case commandElapsed:
		his := gmetrics.NewHistogram(gmetrics.NewUniformSample(defaultReservoirSize))
		if h, ok := m.r.GetOrRegister(p.key, his).(gmetrics.Histogram); ok {
			h.Update(p.elapsed)
		}
	case commandGauge:
		gf := gmetrics.NewGaugeFloat64()
		if h, ok := m.r.GetOrRegister(p.key, gf).(gmetrics.GaugeFloat64); ok {
			h.Update(p.fv)
		}
	case commandRPC:
		his := gmetrics.NewHistogram(gmetrics.NewUniformSample(defaultReservoirSize))
		if h, ok := m.r.GetOrRegister(genRPCKey(p.caller, p.callee), his).(gmetrics.Histogram); ok {
			h.Update(p.elapsed)
		}
	}
}

func (m *metricsStruct) input(p *packet) {
	select {
	case m.in <- p:
	default:
	}
}

// AddOneDeltForMultiKeys ...
func AddOneDeltForMultiKeys(keys ...string) {
	for _, key := range keys {
		Add(key, 1)
	}
}

// Add ...
func Add(key string, delt int64) {
	if defaultClient == nil {
		return
	}
	p := pool.Get().(*packet)
	p.cmd = commandAdd
	p.key = key
	p.c = delt
	defaultClient.input(p)
}

// AddOneDeltForMultiErrorKeys ...
func AddOneDeltForMultiErrorKeys(errCode string, keys ...string) {
	for _, key := range keys {
		AddError(key, errCode, 1)
	}
}

// AddError ...
func AddError(key, errCode string, delt int64) {
	if defaultClient == nil {
		return
	}
	p := pool.Get().(*packet)
	p.cmd = commandAddErr
	p.key = key
	p.code = errCode
	p.c = delt
	defaultClient.input(p)
}

// Elapsed ...
func Elapsed(key string, delay time.Duration) {
	if defaultClient == nil {
		return
	}
	p := pool.Get().(*packet)
	p.cmd = commandElapsed
	p.key = key
	p.elapsed = int64(delay / defaultClient.durationUnit)
	defaultClient.input(p)
}

// GaugeUpdate ...
func GaugeUpdate(key string, fv float64) {
	if defaultClient == nil {
		return
	}
	p := pool.Get().(*packet)
	p.cmd = commandGauge
	p.key = key
	p.fv = fv
	defaultClient.input(p)
}

// RPC ...
func RPC(caller, callee string, elapsed time.Duration, code interface{}) {
	if defaultClient == nil {
		return
	}
	key := genRPCCountKey(caller, callee)
	Add(key, 1)
	strcode, b := isOK(code)
	if defaultClient.debug {
		fmt.Printf("caller=%s callee=%s elapsed=%s code=%v strcode=%s b=%v\n", caller, callee, elapsed, code, strcode, b)
	}
	if b {
		RPCOk(caller, callee, elapsed, strcode)
	} else {
		AddError(key, strcode, 1)
	}
}

func isOK(code interface{}) (string, bool) {
	var strcode string
	var b bool
	switch code.(type) {
	case int:
		strcode = strconv.Itoa(code.(int))
		if code == 0 {
			b = true
		} else {
			b = false
		}
	case int32:
		strcode = fmt.Sprint(code)
		if strcode == "0" {
			b = true
		} else {
			b = false
		}
	case int64:
		strcode = fmt.Sprint(code)
		if strcode == "0" {
			b = true
		} else {
			b = false
		}
	case string:
		strcode = code.(string)
		if code == "0" || code == "OK" || code == "ok" || code == "Ok" {
			b = true
		} else {
			b = false
		}
	default:
		strcode = "unknown"
		b = false

	}
	return strcode, b
}

// RPCOk ...
func RPCOk(caller, callee string, esapled time.Duration, code string) {
	if defaultClient == nil {
		return
	}
	p := pool.Get().(*packet)
	p.cmd = commandRPC
	p.caller = caller
	p.callee = callee
	p.elapsed = int64(esapled / defaultClient.durationUnit)
	p.code = code
	defaultClient.input(p)
}

func genErrKey(key, code string) string {
	return key + "#" + code
}

func parseErrKey(errKey string) (key, code string) {
	index := lastIndexByte(errKey, '#')
	if index > 0 {
		return errKey[:index], errKey[index+1:]
	}
	return errKey, ""
}

func genRPCKey(caller, callee string) string {
	return caller + "@" + callee
}

func parseRPCKey(rpcKey string) (caller, callee string) {
	index := lastIndexByte(rpcKey, '@')
	if index > 0 {
		return rpcKey[:index], rpcKey[index+1:]
	}
	return rpcKey, ""
}

func genRPCCountKey(caller, callee string) string {
	return caller + "$" + callee
}

func parseRPCCountKey(rpcKey string) (caller, callee string) {
	index := lastIndexByte(rpcKey, '$')
	if index > 0 {
		return rpcKey[:index], rpcKey[index+1:]
	}
	return rpcKey, ""
}

// lastIndexByte should be replaced when go_version > 1.6
func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}
