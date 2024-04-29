package rpc2

import "github.com/apache/thrift/lib/go/thrift"

func TestRpc2() {
	serverTransport, _ := thrift.NewTServerSocket("127.0.0.1")
	transportFactory := thrift.NewTFramedTransportFactory(thrift.NewTTransportFactory())
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	server := thrift.NewTSimpleServer4(nil, serverTransport, transportFactory, protocolFactory)
	server.Serve()
}
