//Package metrics ...
package metrics

import (
	"bytes"
	"encoding/json"

	"github.com/dup2X/gopkg/logger"

	gmetrics "github.com/rcrowley/go-metrics"
)

var _ transport = new(printClient)

type printClient struct {
}

func newPrintClient() *printClient {
	return &printClient{}
}

func (pl *printClient) send(service string, sp map[string]interface{}) {
	endpoint := "localhost"
	rd := genRecords(service, sp)
	res := &PrintStruct{
		Service:  service,
		Endpoint: endpoint,
		Record:   rd,
	}
	data := &bytes.Buffer{}
	json.NewEncoder(data).Encode(res)
	logger.Info(logger.DLTagUndefined, data.String())
}

func (pl *printClient) sendOld(service string, sp map[string]interface{}) {
	endpoint := "localhost"
	rets := make(map[string]*Result)
	for k, v := range sp {
		switch inst := v.(type) {
		case gmetrics.Counter:
			key, code := parseErrKey(k)
			if _, ok := rets[key]; !ok {
				rets[key] = &Result{
					key: key,
					Err: make(map[string]int64),
				}
			}
			if code == "" {
				rets[key].Hit += inst.Count()
			} else {
				rets[key].Err[code] += inst.Count()
			}
		case gmetrics.GaugeFloat64:
			key, code := parseErrKey(k)
			if _, ok := rets[key]; !ok {
				rets[key] = &Result{
					key: key,
					Err: make(map[string]int64),
				}
			}
			if code == "" {
				rets[key].Hit += int64(inst.Value())
			} else {
				rets[key].Err[code] += int64(inst.Value())
			}
		}
	}

	res := &PrintStruct{
		Service:  service,
		Endpoint: endpoint,
		Data:     rets,
	}
	data := &bytes.Buffer{}
	json.NewEncoder(data).Encode(res)
	logger.Info(logger.DLTagUndefined, data.String())
}

//PrintStruct :Print Struct
type PrintStruct struct {
	Service  string             `json:"service"`
	Endpoint string             `json:"endpoint"`
	Data     map[string]*Result `json:"data,omitempty"`
	Record   []*Record          `json:"record,omitempty"`
}

//Result :Result struct
type Result struct {
	key string
	Hit int64            `json:"hit"`
	Err map[string]int64 `json:"error"`
}
