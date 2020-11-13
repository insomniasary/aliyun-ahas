package gateway

import (
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-ahas-go-sdk/aliyun"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type AgwHandler interface {
	Handle(request string) (string, error)
}

type RpcMetadata struct {
	ServerName  string
	HandlerName string
	Version     uint32
}

type AgwConfig struct {
	ClientVpcId       string
	ClientIp          string
	ClientProcessFlag string
	GatewayIp         string
	GatewayPort       uint32
	LogDisabled       bool
	// add for tls
	ClientEnv      string
	ClientRegionId string
	TlsFlag        bool
	Timeout        time.Duration
}

type AgwClient struct {
	config      AgwConfig
	initialized bool
	pool        *ConnectionPool
	timeout     uint32
}

var instance *AgwClient
var cLock sync.Mutex
var initOnce sync.Once
var handlers = make(map[string]AgwHandler)

func GetAgwClientInstance() *AgwClient {

	if instance != nil {
		return instance
	}

	cLock.Lock()
	defer cLock.Unlock()

	if instance != nil {
		return instance
	}

	instance = &AgwClient{
		initialized: false,
		pool:        getConnectionPoolInstance(2),
	}

	return instance
}

func (c *AgwClient) Init(config AgwConfig) error {
	if c.initialized {
		return errors.New("dup init")
	}

	if config.ClientIp == "" {
		return errors.New("ip can not be blank")
	}

	if config.ClientProcessFlag == "" {
		return errors.New("processFlag can not be blank")
	}

	if config.ClientVpcId == "" {
		return errors.New("vpcId can not be blank")
	}
	// check or download the cert if not exists
	if config.TlsFlag {
		err := checkOrDownloadCert()
		if err != nil {
			return err
		}
	}
	initOnce.Do(func() {
		c.config = config
		c.timeout = uint32(c.config.Timeout.Milliseconds())
		c.initialized = true
		go runHeartBeatCoroutine(c)
	})

	return nil
}

func (c *AgwClient) Call(outerReqId string, rpcMetadata RpcMetadata, jsonParam string) (string, error) {
	if !c.initialized {
		return "", errors.New("the client has not be initialized")
	}

	if outerReqId == "" {
		return "", errors.New("reqId can not be blank")
	}

	tsUtil := newTimestampUtilV2(outerReqId, c.config.ClientVpcId, c.config.ClientProcessFlag, c.config.ClientIp)
	tsUtil.mark("client_call_gateway")

	var response *AgwMessage
	var responseError error
	var reqId uint64
	for retryTime := default_req_retry_time; retryTime > 0; retryTime-- {
		reqId = generateId()
		response, responseError = c.innerCall(reqId, outerReqId, rpcMetadata, jsonParam)
		if responseError == nil {
			break
		}
		errMsg := responseError.Error()
		if strings.Compare(errMsg, ErrorMsgRequestTimeout) == 0 {
			logWarnf("gateway retry for timeout, reqId:%d, outerReqId:%s", reqId, outerReqId)
			continue
		}
		if strings.Compare(errMsg, ErrorMsgConnClosed) == 0 {
			logWarnf("gateway retry for connection close, reqId:%d, outerReqId:%s", reqId, outerReqId)
			continue
		}
		if strings.Contains(errMsg, ErrorMsgWriteClosedConn) {
			logWarnf("gateway retry for broken pipe, reqId:%d, outerReqId:%s", reqId, outerReqId)
			continue
		}
		if strings.Contains(errMsg, ErrorMsgUseClosedConn) {
			logWarnf("gateway retry for using closed connection, reqId:%d, outerReqId:%s", reqId, outerReqId)
			continue
		}

		break
	}

	tsUtil.SetReqId(reqId)
	if responseError != nil {
		logWarnf("a net error happens after some times of retry, reqId:%d, outerReqId:%s", reqId, outerReqId)
		tsUtil.mark("rpc_error")
		logDebug(tsUtil.GetResultV2())
		return "", responseError
	}

	if response.InnerCode() != 0 {
		logWarnf("a biz error happens, reqId:%d, outerReqId:%s", reqId, outerReqId)
		tsUtil.mark("biz_error")
		logDebug(tsUtil.GetResultV2())
		return "", errors.New(fmt.Sprintf("call error [%d:%s]", response.InnerCode(), response.InnerMsg()))
	}

	tsUtil.mark("after_call")
	logDebug(tsUtil.GetResultV2())
	return response.Body(), nil
}

func (c *AgwClient) innerCall(reqId uint64, outerReqId string, rpcMetadata RpcMetadata, jsonParam string) (*AgwMessage, error) {
	conn, err := c.pool.get()
	if err != nil {
		return nil, err
	}

	msg := NewAgwMessage()
	msg.SetReqId(reqId)
	msg.SetMessageType(MessageTypeBiz)
	msg.SetMessageDirection(MessageDirectionRequest)
	msg.SetClientIp(StringIpToUint64(c.config.ClientIp))
	msg.SetClientVpcId(c.config.ClientVpcId)
	msg.SetServerName(rpcMetadata.ServerName)
	msg.SetTimeoutMs(c.timeout)
	msg.SetClientProcessFlag(c.config.ClientProcessFlag)
	msg.SetConnectionId(conn.connId)
	msg.SetHandlerName(rpcMetadata.HandlerName)
	msg.SetOuterReqId(outerReqId)
	msg.SetBody(jsonParam)
	msg.SetVersion(rpcMetadata.Version)

	return conn.writeSync(msg)
}

func (c *AgwClient) AddHandler(handlerName string, handler AgwHandler) error {
	if handlerName == "" {
		return errors.New("handlerName can not be blank")
	}

	if handler == nil {
		return errors.New("hander can not be null")
	}

	logInfof("Adding handler to AgwClient: %s", handlerName)
	handlers[handlerName] = handler

	return nil
}

var CertPath = path.Join(os.TempDir(), ".server.cert")

func checkOrDownloadCert() error {
	if tools.IsExist(CertPath) {
		return nil
	}
	remoteFilePath := path.Join(tools.Constant.OSAgentRemotePath, "cert", "sChat.pem")
	err := aliyun.Download(CertPath, meta.RegionId(), remoteFilePath, meta.IsPrivate())
	if err != nil {
		return fmt.Errorf("download cert failed, %v", err)
	}
	return nil
}
