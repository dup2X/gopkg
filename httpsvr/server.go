// Package httpsvr ...
package httpsvr

import (
	"bytes"
	stdctx "context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/dup2X/gopkg/context"
	"github.com/dup2X/gopkg/elapsed"
	"github.com/dup2X/gopkg/idl"
	"github.com/dup2X/gopkg/logger"
	"github.com/dup2X/gopkg/metrics"
	"github.com/dup2X/gopkg/utils"
	"github.com/julienschmidt/httprouter"
)

//Server 定义Server结构体
type Server struct {
	addr   string
	router *httprouter.Router
	log    logger.Logger
	mid    []Middleware
	opt    *option
	oriSvr *http.Server
}

//New 创建默认Server
func New(addr string, opts ...ServerOption) *Server {
	opt := &option{}
	for _, o := range opts {
		o(opt)
	}
	if addr == "" {
		addr = "127.0.0.1:10024"
	}
	if opt.marshalFunc == nil {
		opt.marshalFunc = defaultMarshalFn
	}
	if opt.unmarshalFunc == nil {
		opt.unmarshalFunc = jsonDecodeFunc
	}
	if opt.handleTimeout == 0 {
		opt.handleTimeout = int64(defaultTimeout)
	}
	s := &Server{
		addr:   addr,
		router: httprouter.New(),
		opt:    opt,
	}
	s.oriSvr = &http.Server{Addr: addr, Handler: s}
	if opt.readTimeout > 0 {
		s.oriSvr.ReadTimeout = opt.readTimeout
	}
	if opt.writeTimeout > 0 {
		s.oriSvr.WriteTimeout = opt.writeTimeout
	}
	return s
}

//ServeHTTP server http
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

//AddRoute 添加路由
func (s *Server) AddRoute(method, path string, ctrl idl.IController, opts ...ControllerOption) {
	var proc httprouter.Handle = func(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
		et := elapsed.New()
		et.Start()
		cos := &ctrlOption{}
		for _, o := range opts {
			o(cos)
		}
		adp := newHTTPAdapter(cos, s.opt)
		if s.opt.dumpAccess {
			if s.log != nil {
				s.log.Infof(logger.DLTagRequestIn+"||uri=%s||query=%s", r.URL.Path, r.URL.RawQuery)
			} else {
				ctx := stdctx.TODO()
				ctx = setTrace(ctx, r)
				ct, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if ct == "multipart/form-data" {
					r.ParseMultipartForm(10 << 20)
					paramStr := ""
					for key := range r.PostForm {
						if paramStr != "" {
							paramStr = paramStr + "&"
						}
						paramStr = paramStr + fmt.Sprintf("%s=%s", key, r.Form.Get(key))
					}
					logger.Infof(ctx, logger.DLTagRequestIn, "uri=%s||client_ip=%s||request_body=%s",
						r.URL,
						utils.GetClientAddr(r),
						paramStr)
				} else {
					body, _ := ioutil.ReadAll(r.Body)
					r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
					logger.Infof(ctx, logger.DLTagRequestIn, "uri=%s||client_ip=%s||request_body=%s",
						r.URL,
						utils.GetClientAddr(r),
						string(body))
				}

			}
		}
		def := ctrl.GetRequestIDL()
		ctx, err := adp.Accept(r, def)
		if err != nil {
			metrics.Add(bindInputParamFailed, 1)
			w.WriteHeader(http.StatusOK)
			w.Write(getErrMsg(err))
			return
		}

		do := func(ctx stdctx.Context, r *http.Request, w http.ResponseWriter) {
			var data []byte
			resp, code := ctrl.Do(ctx, def)
			if resp == nil {
				data = defaultResponse
			}
			if adp.marshalFunc != nil {
				data, _ = adp.marshalFunc(resp, code)
			}
			if s.opt.dumpResponse {
				s.log.Trace(utils.DumpHex(data))
			}
			et.Stop()
			if code == nil {
				code = defaultOk
			}
			if s.log != nil {
				s.log.Infof(logger.DLTagRequestOut+"||%s||response=%s", ctx, string(data))
			} else {
				logger.Infof(ctx, logger.DLTagRequestOut,
					"uri=%s||response=%s||errno=%d||errmsg=%s||redis_elapsed=%s||mysql_elapsed=%s||proc_time=%d",
					r.URL, string(data), code.Code(), code.Error(), context.GetRedisElapsed(ctx), context.GetMysqlElapsed(ctx), et.Elapsed()/1e6)
			}
			w.WriteHeader(200)
			w.Write(data)
		}
		if adp.addResponseHeader != nil {
			hd := adp.addResponseHeader()
			for k, vs := range hd {
				for i := range vs {
					w.Header().Add(k, vs[i])
				}
			}
		}
		if !context.CheckSLA(ctx) {
			w.Write(shortSLAResponse)
			w.WriteHeader(200)
			return
		}
		if s.mid != nil {
			for idx := len(s.mid) - 1; idx >= 0; idx-- {
				do = s.mid[idx](ctx, r, w, do)
			}
		}
		do(ctx, r, w)
	}
	s.router.Handle(method, path, proc)
}

//AddMiddleware 添加middleware
func (s *Server) AddMiddleware(md Middleware) {
	if s.mid == nil {
		s.mid = make([]Middleware, 0)
	}
	s.mid = append(s.mid, md)
}

func (s *Server) HandleFunc(method, path string, hd func(w http.ResponseWriter, r *http.Request)) {
	s.router.HandlerFunc(method, path, hd)
}

//Serve 监听普通连接
func (s *Server) Serve() error {
	return s.oriSvr.ListenAndServe()
}

//ServeTLS 监听SSL连接
func (s *Server) ServeTLS(certFile, keyFile string) error {
	return s.oriSvr.ListenAndServeTLS(certFile, keyFile)
}

func getErrMsg(err error) []byte {
	return []byte(fmt.Sprintf(`{"errno":-1,"errmsg":"%s"}`, err.Error()))
}

var (
	defaultResponse = []byte(`{"errno":0,"errmsg":"ok"}`)
	localLanguage   string
	defaultOk       = &okErr{}
)

var jsonDecodeFunc = func(r *http.Request, req interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	return json.Unmarshal(data, req)
}

const (
	bindInputParamFailed = "http_framework_parse_parameters_failed"
)
