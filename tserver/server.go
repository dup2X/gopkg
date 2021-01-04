package tserver

import (
	"errors"
	"sync"

	"git.apache.org/thrift.git/lib/go/thrift"
)

var (
	errUninit = errors.New("please init server before start")

	once = sync.Once{}
)

// Server defined thrift server
type Server struct {
	svrAddd   string
	processor thrift.TProcessor
}

// NewServer return Server
func NewServer(addr string, processor thrift.TProcessor) *Server {
	return &Server{
		svrAddd:   addr,
		processor: processor,
	}
}

// Serve define main
func (s *Server) Serve() error {
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	serverTransport, err := thrift.NewTServerSocket(s.svrAddd)
	if err != nil {
		return err
	}
	server := thrift.NewTSimpleServer4(s.processor, serverTransport, transportFactory, protocolFactory)
	return server.Serve()
}

// Stop defined
func (s *Server) Stop() error {
	once.Do(func() {})
	return nil
}
