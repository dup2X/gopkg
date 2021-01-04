//Package metrics ...
package metrics

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/dup2X/gopkg/config"
	"github.com/dup2X/gopkg/logger"
)

var mockOdinServer = func() {
	http.HandleFunc("/v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		println("URL =", r.URL.String())
		data, _ := ioutil.ReadAll(r.Body)
		println("DATA =", string(data))
	})
	err := http.ListenAndServe(":10791", nil)
	if err != nil {
		println(err.Error())
	}
}

func TestAdd(t *testing.T) {
	cfg, err := config.New("./testdata/test.conf")
	if err != nil {
		t.FailNow()
	}
	err = logger.NewLoggerWithConfig(cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	sec, err := cfg.GetSection("metrics")
	if err != nil {
		t.FailNow()
	}
	err = NewWithConfig(sec)
	if err != nil {
		t.FailNow()
	}
	go mockOdinServer()
	Add("k1", 1)
	Add("k1", 1)
	Add("k2", 1)
	AddError("k1", "401", 1)
	AddError("k1", "20132", 1)
	Elapsed("k2", time.Now().Sub(time.Now().Add(-1*time.Second)))
	GaugeUpdate("number--------", 12.0)
	time.Sleep(time.Second * 2)
}

func TestRPC(t *testing.T) {
	cfg, err := config.New("./testdata/test.conf")
	if err != nil {
		t.FailNow()
	}
	err = logger.NewLoggerWithConfig(cfg)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	sec, err := cfg.GetSection("metrics")
	if err != nil {
		t.FailNow()
	}
	err = NewWithConfig(sec)
	if err != nil {
		t.FailNow()
	}
	//go mockOdinServer()
	for i := 0; i < 100; i++ {
		RPC("caller1", "callee-a", time.Millisecond*time.Duration(rand.Int63n(400)), 0)
		RPC("caller1", "callee-a", time.Millisecond*time.Duration(rand.Int63n(200)), 10023)
		RPC("caller1", "callee-a", time.Millisecond*time.Duration(rand.Int63n(40)), 404)
		RPC("caller2", "callee-a2", time.Millisecond*time.Duration(rand.Int63n(40)), int64(0))
		time.Sleep(time.Millisecond * 10)
	}
	time.Sleep(time.Second * 2)
}
