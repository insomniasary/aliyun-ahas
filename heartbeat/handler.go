package heartbeat

import (
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
)

type Handler struct {
	transport.AgwRequestHandler
}

func GetPingHandler() *Handler {
	handler := &Handler{}
	pingHandler := transport.NewCommonHandler(handler)
	handler.AgwRequestHandler = pingHandler
	return handler
}

func (handler *Handler) Handle(request *transport.Request) *transport.Response {
	return transport.ReturnSuccess("success")
}
