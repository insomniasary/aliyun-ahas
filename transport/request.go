package transport

import (
	"fmt"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"

	"github.com/aliyun/aliyun-ahas-go-sdk/gateway"
)

const (
	FromHeader = "FR"
	Client     = "C"
	Aid        = "aid"
	Tid        = "tid"
	Pid        = "pid"
	Uid        = "uid"
)

var (
	NoCompress  = fmt.Sprintf("%d", gateway.NoCompress)
	AllCompress = fmt.Sprintf("%d", gateway.AllCompress)
)

type Request struct {
	Headers map[string]string `json:"headers"`
	Params  map[string]string `json:"params"`
}

func NewRequest() *Request {
	request := &Request{
		Headers: make(map[string]string),
		Params:  make(map[string]string),
	}
	request.AddHeader(FromHeader, Client)
	return request
}

// AddHeader add metadata to it
func (request *Request) AddHeader(key string, value string) *Request {
	if key != "" {
		request.Headers[key] = value
	}
	return request
}

// AddParam add request data to it
func (request *Request) AddParam(key string, value string) *Request {
	if key != "" {
		request.Params[key] = value
	}
	return request
}

const DELIMITER = "_"

const (
	// Topology service
	Topology  = "Topology"
	Connect   = "connect"
	Heartbeat = "heartbeat"
	Close     = "close"

	// Client service
	Ping = "ping"
)

type Uri struct {
	ServerName      string
	HandlerName     string
	VpcId           string
	Ip              string
	Pid             string
	Tag             string
	RequestId       string
	CompressVersion string
}

// NewUri: create a new one
func NewUri(serverName, handlerName string) Uri {
	return Uri{
		ServerName:      serverName,
		HandlerName:     handlerName,
		VpcId:           meta.VpcId(),
		Ip:              meta.LocalIp(),
		Pid:             meta.Pid(),
		Tag:             meta.GoSDK,
		CompressVersion: NoCompress,
	}
}
