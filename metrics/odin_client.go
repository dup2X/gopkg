//Package metrics ...
package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"github.com/dup2X/gopkg/logger"

	gmetrics "github.com/rcrowley/go-metrics"
)

var (
	defaultPercents = []float64{0.50, 0.75, 0.95, 0.99}

	_ transport = new(odinClient)
)

type odinClient struct {
	reportURL   string
	sendTimeout time.Duration
	cli         *http.Client
}

func newOdinClient(reqURL string, timeout time.Duration) *odinClient {
	return &odinClient{
		reportURL:   reqURL,
		sendTimeout: timeout,
		cli: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					deadline := time.Now().Add(timeout)
					c, err := net.DialTimeout(netw, addr, timeout)
					if err != nil {
						return nil, err
					}
					c.SetDeadline(deadline)
					return c, nil
				},
			},
		},
	}
}

func (od *odinClient) send(service string, snapshot map[string]interface{}) {
	if len(snapshot) == 0 {
		return
	}
	rs := genRecords(service, snapshot)
	data, err := json.Marshal(rs)
	if err != nil {
		logger.Error(logger.DLTagUndefined, err)
		return
	}
	if defaultClient.debug {
		fmt.Printf("======odin ===== %s\n", string(data))
	}
	req, err := http.NewRequest("POST", od.reportURL, bytes.NewReader(data))
	if err != nil {
		logger.Error(logger.DLTagUndefined, err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := od.cli.Do(req)
	if err != nil {
		// TODO retry
		logger.Error(logger.DLTagUndefined, err)
		return
	}
	resp.Body.Close()
}

// Record ...
type Record struct {
	Name  string      `json:"name"`
	Tags  Tag         `json:"tags"`
	Ts    int64       `json:"timestamp"`
	Value interface{} `json:"value"`
	Step  int64       `json:"step"`
}

// Tag ...
type Tag struct {
	Service string `json:"service,omitempty"`
	Key     string `json:"api,omitempty"`
	Code    string `json:"code,omitempty"`
	Caller  string `json:"caller,omitempty"`
	Callee  string `json:"callee,omitempty"`
	Percent string `json:"percentile,omitempty"`
}

type pair struct {
	hit int64
	err int64
}

func genRecords(service string, ss map[string]interface{}) []*Record {
	var rs []*Record
	count := make(map[string]*pair)

	ts := time.Now().Unix()
	for k, v := range ss {
		switch inst := v.(type) {
		case gmetrics.Counter:
			val := inst.Count()
			r := &Record{
				Ts:    ts,
				Value: val,
				Step:  reportTTL,
			}
			key, code := parseErrKey(k)
			if _, ok := count[key]; !ok {
				count[key] = &pair{}
			}
			caller, callee := parseRPCCountKey(key)
			if callee == "" {
				r.Tags = Tag{
					Service: service,
					Key:     key,
					Code:    code}
				if code != "" {
					r.Name = "error_count"
					count[key].err += val
				} else {
					r.Name = "query_count"
					count[key].hit += val
				}
				rs = append(rs, r)
			} else {
				r.Tags = Tag{
					Service: service,
					Caller:  caller,
					Callee:  callee,
					Code:    code}
				if code != "" {
					r.Name = "rpc.error.counter"
					count[key].err += val
				} else {
					r.Name = "rpc.counter"
					count[key].hit += val
				}
				rs = append(rs, r)
			}
		case gmetrics.GaugeFloat64:
			val := inst.Value()
			r := &Record{
				Ts:    ts,
				Value: val,
				Step:  reportTTL,
			}

			r.Tags = Tag{
				Service: service,
				Key:     k}
			r.Name = "query_count"
			rs = append(rs, r)
		case gmetrics.Histogram:
			_, callee := parseRPCKey(k)
			if callee == "" {
				rs = append(rs, genHist(service, ts, k, inst)...)
			} else {
				rs = append(rs, genRPCHist(service, ts, k, inst)...)
			}
		}
	}
	rs = append(rs, errorRate(service, ts, count)...)
	return rs
}

func genHist(service string, ts int64, k string, inst gmetrics.Histogram) []*Record {
	var rs []*Record
	vals := inst.Percentiles(defaultPercents)
	for i, pc := range defaultPercents {
		rs = append(rs, &Record{
			Name: fmt.Sprintf("latency_%dth", int(100*pc)),
			Ts:   ts,
			Tags: Tag{
				Service: service,
				Key:     k},
			Value: vals[i],
			Step:  reportTTL,
		})
	}
	rs = append(rs, &Record{
		Name: "latency_avg",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Key:     k},
		Value: inst.Mean(),
		Step:  reportTTL,
	})
	rs = append(rs, &Record{
		Name: "latency_max",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Key:     k},
		Value: inst.Max(),
		Step:  reportTTL,
	})
	rs = append(rs, &Record{
		Name: "latency_min",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Key:     k},
		Value: inst.Min(),
		Step:  reportTTL,
	})
	return rs
}

//rpc调用的延时百分位统计
func genRPCHist(service string, ts int64, k string, inst gmetrics.Histogram) []*Record {
	var rs []*Record
	vals := inst.Percentiles(defaultPercents)
	caller, callee := parseRPCKey(k)
	for i, pc := range defaultPercents {
		rs = append(rs, &Record{
			Name: "rpc.latency",
			Ts:   ts,
			Tags: Tag{
				Service: service,
				Caller:  caller,
				Callee:  callee,
				Percent: fmt.Sprint(int(100 * pc)),
			},
			Value: vals[i],
			Step:  reportTTL,
		})
	}
	rs = append(rs, &Record{
		Name: "rpc.latency",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Caller:  caller,
			Callee:  callee,
			Percent: "avg",
		},
		Value: inst.Mean(),
		Step:  reportTTL,
	})
	rs = append(rs, &Record{
		Name: "rpc.latency",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Caller:  caller,
			Callee:  callee,
			Percent: "max",
		},
		Value: inst.Max(),
		Step:  reportTTL,
	})
	rs = append(rs, &Record{
		Name: "rpc.latency",
		Ts:   ts,
		Tags: Tag{
			Service: service,
			Caller:  caller,
			Callee:  callee,
			Percent: "min",
		},
		Value: inst.Min(),
		Step:  reportTTL,
	})
	return rs
}

//统计上报数据的错误率
func errorRate(service string, ts int64, count map[string]*pair) []*Record {
	var rs []*Record
	for k, p := range count {
		val := float64(p.err) / float64(p.hit)
		if math.IsNaN(val) || math.IsInf(val, 1) || math.IsInf(val, -1) {
			continue
		}
		caller, callee := parseRPCCountKey(k)
		valPercent := float64(int(val*100)) / 100.00
		if callee == "" {
			rs = append(rs, &Record{
				Name: "error_rate",
				Ts:   ts,
				Tags: Tag{
					Service: service,
					Key:     k,
				},
				Value: valPercent,
				Step:  reportTTL,
			})
		} else {
			rs = append(rs, &Record{
				Name: "rpc.error.ratio",
				Ts:   ts,
				Tags: Tag{
					Service: service,
					Caller:  caller,
					Callee:  callee},
				Value: valPercent,
				Step:  reportTTL,
			})

		}
	}
	return rs
}
