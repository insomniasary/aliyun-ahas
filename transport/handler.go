package transport

import (
	"encoding/json"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"github.com/aliyun/aliyun-ahas-go-sdk/service"
)

type RequestHandler interface {
	Handle(request *Request) *Response
}

type AgwRequestHandler struct {
	Interceptor RequestInterceptor
	Handler     RequestHandler
	*service.Controller
}

func (handler *AgwRequestHandler) DoStart() error {
	return nil
}

func (handler *AgwRequestHandler) DoStop() error {
	return nil
}

//NewCommonHandler with default interceptor
func NewCommonHandler(handler RequestHandler) AgwRequestHandler {
	requestHandler := AgwRequestHandler{
		Interceptor: buildInterceptor(),
		Handler:     handler,
	}
	requestHandler.Controller = service.NewController(&requestHandler)
	requestHandler.Start()
	return requestHandler
}

func (handler *AgwRequestHandler) Handle(request string) (string, error) {
	var response *Response = nil
	select {
	case <-handler.Ctx.Done():
		response = ReturnFail(Code[HandlerClosed], Code[HandlerClosed].Msg)
	default:
		// decode
		req := &Request{}
		err := json.Unmarshal([]byte(request), req)
		if err != nil {
			return "", err
		}
		var ok = true
		// interceptor
		interceptor := handler.Interceptor
		if interceptor != nil && !meta.DebugEnabled() {
			response, ok = interceptor.Handle(req)
		}
		if ok {
			// Call Handler only when passing the interceptor
			response = handler.Handler.Handle(req)
		}
	}
	// encode
	bytes, err := json.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
