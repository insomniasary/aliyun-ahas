package transport

import (
	"errors"
	"fmt"
	sentinelConf "github.com/alibaba/sentinel-golang/core/config"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-ahas-go-sdk/gateway"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
)

type Transport struct {
	client   *gateway.AgwClient
	invoker  RequestInvoker
	handlers map[string]*AgwRequestHandler
	mutex    sync.Mutex
	config   *Config
	metadata *meta.Meta
}

func (t *Transport) Shutdown() error {
	return nil
}

func New(conf *Config, metadata *meta.Meta) (*Transport, error) {
	if conf == nil {
		return nil, errors.New("nil transport config")
	}
	if metadata == nil {
		return nil, errors.New("nil metadata")
	}
	client := gateway.GetAgwClientInstance()

	hostAndPort := strings.SplitN(metadata.AhasEndpoint(), ":", 2)
	port, err := strconv.Atoi(hostAndPort[1])
	if err != nil {
		return nil, err
	}
	// TODO: privateIp
	ip := metadata.Ip()
	processFlag := meta.GoSDK + ":" + ip + ":" + metadata.Pid()

	if conf.TimeoutMs == 0 {
		conf.TimeoutMs = 3000
	}
	agwConfig := gateway.AgwConfig{
		ClientVpcId:       metadata.VpcId(),
		ClientIp:          ip,
		ClientProcessFlag: processFlag,
		GatewayIp:         hostAndPort[0],
		GatewayPort:       uint32(port),
		Timeout:           time.Duration(conf.TimeoutMs) * time.Millisecond,
	}
	secure := true
	if secure {
		agwConfig.ClientRegionId = metadata.RegionId()
		agwConfig.ClientEnv = meta.DeployEnv()
		agwConfig.TlsFlag = true
	}
	err = client.Init(agwConfig)
	if err != nil {
		return nil, err
	}
	return &Transport{
		client:   client,
		invoker:  NewInvoker(client, true),
		handlers: make(map[string]*AgwRequestHandler),
		mutex:    sync.Mutex{},
		config:   conf,
		metadata: metadata,
	}, nil
}

//addHandler register handler
func (t *Transport) RegisterHandler(handlerName string, handler *AgwRequestHandler) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.handlers[handlerName] == nil {
		t.handlers[handlerName] = handler
		t.client.AddHandler(handlerName, handler)
	}
	if meta.DebugEnabled() {
		http.HandleFunc("/ahas/"+handlerName, func(writer http.ResponseWriter, request *http.Request) {
			request.ParseForm()
			response, err := handler.Handle(request.Form["body"][0])
			if err != nil {
				response = err.Error()
			}
			fmt.Fprintf(writer, response)
		})
	}
}

//Start Transport service
func (t *Transport) Start() (*Transport, error) {
	err := t.connect()
	if err != nil {
		logger.Errorf("Connection to server failed: %+v", err)
		return nil, err
	}
	logger.Info("AGW transport service started successfully")
	return t, nil
}

func (t *Transport) Stop() error {
	return nil
}

// Connect to remote
func (t *Transport) connect() error {
	// TODO
	request := NewRequest()
	request.AddParam("vpcId", t.metadata.VpcId()).AddParam("ip", t.metadata.Ip())
	request.AddParam("pid", t.metadata.Pid()).AddParam("type", meta.GoSDK)
	request.AddParam("appName", sentinelConf.AppName())
	request.AddParam("appType", strconv.Itoa(int(sentinelConf.AppType())))
	request.AddParam("namespace", meta.Namespace())

	uid := t.metadata.Uid()
	if uid == "" {
		request.AddParam("ak", meta.License())
	} else {
		request.AddParam("uid", uid)
	}

	request.AddParam("deviceId", t.metadata.InstanceId())
	request.AddParam("deviceType", strconv.Itoa(meta.Host))

	request.AddParam("v", t.metadata.Version())
	request.AddParam("hostIp", t.metadata.Ip())
	request.AddParam("cpuNum", strconv.Itoa(runtime.NumCPU()))

	uri := NewUri(Topology, Connect)
	invoker := NewInvoker(t.client, false)
	response, err := invoker.Invoke(uri, request)
	if err != nil {
		return err
	}
	return handleConnectResponse(*response, t.metadata)
}

// Handle response: record ak/sk and uid information
func handleConnectResponse(response Response, metadata *meta.Meta) error {
	if !response.Success {
		if response.Code == Code[ServiceNotOpened].Code {
			logger.Errorf("AHAS service not opened, please initiate it in the AHAS console")
		} else if response.Code == Code[ServiceNotAuthorized].Code {
			logger.Errorf("AHAS service not authorized")
		}
		return errors.New(fmt.Sprintf("connect server failed, %s", response.Error))
	}
	result := response.Result

	v, ok := result.(map[string]interface{})
	if !ok {
		return errors.New("response is error")
	}
	if v[Tid] == nil || v[Tid] == "" {
		return errors.New("tid is empty")
	}
	if v[Uid] == nil || v[Uid] == "" {
		return errors.New("uid is empty")
	}

	metadata.SetUid(v[Uid].(string))
	metadata.SetTid(v[Tid].(string))
	metadata.SetCid(v[Aid].(string))

	err := tools.SaveMetadataToFile(v["ak"].(string), v["sk"].(string))
	return err
}

// Invoke remote service. Client communicates with server through this interface
func (t *Transport) Invoke(uri Uri, request *Request) (*Response, error) {
	request.AddHeader(Pid, t.metadata.Pid())
	uid := t.metadata.Uid()
	if uid != "" {
		request.AddHeader(Uid, uid)
	}

	request.AddHeader(Aid, t.metadata.Cid())

	request.AddHeader("type", meta.GoSDK)
	request.AddHeader("v", t.metadata.Version())
	return t.invoker.Invoke(uri, request)
}
