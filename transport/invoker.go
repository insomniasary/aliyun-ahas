package transport

import (
	"encoding/json"
	"github.com/aliyun/aliyun-ahas-go-sdk/gateway"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"github.com/pkg/errors"
	"strconv"
)

//RequestInvoker invoke remote service and return response
type RequestInvoker interface {
	Invoke(uri Uri, request *Request) (*Response, error)
}

type doRequestInvoker interface {
	doInvoker(uri Uri, jsonParam string) (string, error)
}

// invoker with interceptor
type agwRequestInvoker struct {
	interceptor RequestInterceptor
	RequestInvoker
	doRequestInvoker
}

func (invoker *agwRequestInvoker) Invoke(uri Uri, request *Request) (*Response, error) {
	// interceptor
	interceptor := invoker.interceptor
	if interceptor != nil {
		if response, ok := interceptor.Invoke(request); !ok {
			return response, errors.New(response.Error)
		}
	}

	// set requestId
	var requestId = tools.GetUUID()
	request.AddHeader("rid", requestId)
	uri.RequestId = requestId

	// encode
	bytes, err := json.Marshal(request)
	if err != nil {
		logger.Warnf("Marshal request to json error (%s, %s): %+v", uri.ServerName, uri.HandlerName, err)
		return nil, err
	}
	// doInvoke
	result, err := invoker.doInvoker(uri, string(bytes))
	if err != nil {
		logger.Warnf("Invoke failed, requestId: %s, error: %s", requestId, err.Error())
		return nil, err
	}
	// decode
	var response Response
	err = json.Unmarshal([]byte(result), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Invoke gateway client
type agwClientRequestInvoker struct {
	// gateway client
	client *gateway.AgwClient
	agwRequestInvoker
}

func NewInvoker(client *gateway.AgwClient, needInterceptor bool) RequestInvoker {
	//  Not need request interceptor when first connect,
	var interceptor RequestInterceptor
	if needInterceptor {
		interceptor = buildInterceptor()
	} else {
		interceptor = nil
	}
	// entry invoker
	invoker := &agwClientRequestInvoker{
		client,
		agwRequestInvoker{
			interceptor: interceptor,
		},
	}
	invoker.doRequestInvoker = invoker
	invoker.RequestInvoker = invoker
	return invoker
}

func buildInterceptor() RequestInterceptor {
	// auth
	authInterceptor := &authInterceptor{}
	chain := requestInterceptorChain{}
	chain.chain = nil
	chain.RequestInterceptor = &chain
	chain.doRequestInterceptor = authInterceptor
	authInterceptor.requestInterceptorChain = chain

	// timestamp
	timestampInterceptor := &timestampInterceptor{}
	timeChain := requestInterceptorChain{}
	timeChain.chain = authInterceptor
	timeChain.RequestInterceptor = &timeChain
	timeChain.doRequestInterceptor = timestampInterceptor
	timestampInterceptor.requestInterceptorChain = timeChain

	return timestampInterceptor
}

func (invoker *agwClientRequestInvoker) doInvoker(uri Uri, jsonParam string) (string, error) {
	ver, err := strconv.Atoi(uri.CompressVersion)
	if err != nil {
		ver = gateway.AllCompress
	}
	metadata := gateway.RpcMetadata{
		ServerName:  uri.ServerName,
		HandlerName: uri.HandlerName,
		Version:     uint32(ver),
	}
	return invoker.client.Call(uri.RequestId, metadata, jsonParam)
}
