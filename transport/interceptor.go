package transport

import (
	"encoding/json"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"strconv"
	"time"
)

const (
	SignData       = "sd"
	SoleilKey      = "ak"
	SignKey        = "sn"
	TimestampKey   = "ts"
	MaxInvalidTime = 60 * 1000 * time.Millisecond
)

type RequestInterceptor interface {
	Handle(request *Request) (*Response, bool)
	Invoke(request *Request) (*Response, bool)
}

type doRequestInterceptor interface {
	doHandler(request *Request) (*Response, bool)
	doInvoker(request *Request) (*Response, bool)
}

type requestInterceptorChain struct {
	chain RequestInterceptor
	RequestInterceptor
	doRequestInterceptor
}

//Handle interceptor. return nil,true if passed, otherwise return response of fail and false
func (interceptor *requestInterceptorChain) Handle(request *Request) (*Response, bool) {
	if response, ok := interceptor.doHandler(request); !ok {
		return response, ok
	}

	if interceptor != nil && interceptor.chain != nil {
		if response, ok := interceptor.chain.Handle(request); !ok {
			return response, ok
		}
	}
	return nil, true
}

//Invoke interceptor.
func (interceptor *requestInterceptorChain) Invoke(request *Request) (*Response, bool) {
	if response, ok := interceptor.doInvoker(request); !ok {
		return response, ok
	}
	if interceptor.chain != nil {
		if response, ok := interceptor.chain.Invoke(request); !ok {
			return response, ok
		}
	}
	return nil, true
}

type authInterceptor struct {
	requestInterceptorChain
}

func (authInterceptor *authInterceptor) doHandler(request *Request) (*Response, bool) {
	// check sign
	sign := request.Headers[SignKey]
	if sign == "" {
		return ReturnFail(Code[Forbidden], "missing sign"), false
	}
	soleilKey := request.Headers[SoleilKey]
	if soleilKey != "" && soleilKey != tools.GetSoleilKey() {
		return ReturnFail(Code[Forbidden], "soleilKey not matched"), false
	}
	signData := request.Headers[SignData]
	if signData == "" {
		bytes, err := json.Marshal(request.Params)
		if err != nil {
			return ReturnFail(Code[Forbidden], "invalid request parameters"), false
		}
		signData = string(bytes)
	}
	if !tools.Auth(sign, signData) {
		return ReturnFail(Code[Forbidden], "illegal request"), false
	}
	return nil, true
}

func (authInterceptor *authInterceptor) doInvoker(request *Request) (*Response, bool) {
	soleilKey := tools.GetSoleilKey()
	luneKey := tools.GetLuneKey()
	if soleilKey == "" || luneKey == "" {
		return ReturnFail(Code[TokenNotFound], "soleilKey or luneKey not found"), false
	}
	request.AddHeader(SoleilKey, soleilKey)
	signData := request.Headers[SignData]
	if signData == "" {
		bytes, err := json.Marshal(request.Params)
		if err != nil {
			return ReturnFail(Code[EncodeError], err.Error()), false
		}
		signData = string(bytes)
	}
	sign := tools.Sign(signData)
	request.AddHeader(SignKey, sign)
	return nil, true
}

type timestampInterceptor struct {
	requestInterceptorChain
}

func (interceptor *timestampInterceptor) doHandler(request *Request) (*Response, bool) {
	// check timestamp
	requestTime := request.Params[TimestampKey]
	if requestTime == "" {
		return ReturnFail(Code[InvalidTimestamp], Code[InvalidTimestamp].Msg), false
	}
	_, err := strconv.ParseInt(requestTime, 10, 64)
	if err != nil {
		return ReturnFail(Code[InvalidTimestamp], err.Error()), false
	}
	//if getCurrentTimeInMillis()-t > int64(MaxInvalidTime) {
	//	return ReturnFail(Code[Timeout], Code[Timeout].Msg), false
	//}
	return nil, true
}

func (interceptor *timestampInterceptor) doInvoker(request *Request) (*Response, bool) {
	// add timestamp
	currTime := getCurrentTimeInMillis()
	request.AddParam(TimestampKey, strconv.FormatInt(currTime, 10))
	return nil, true
}

func getCurrentTimeInMillis() int64 {
	return time.Now().UnixNano() / 1000
}
