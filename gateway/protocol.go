package gateway

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"runtime/debug"
)

const (
	MessageTypeHeartbeat = 1
	MessageTypeBiz       = 2

	MessageDirectionRequest  = 1
	MessageDirectionResponse = 2
)

const (
	NoCompress       = 1
	AllCompress      = 2
	RequestCompress  = 3
	ResponseCompress = 4
)

type AgwMessage struct {
	tsUtil *timestampUtil

	bodyLength uint32
	reqId uint64
	messageType uint8
	messageDirection uint8
	caller uint8
	clientIp uint64
	clientVpcIdLength uint32
	clientVpcId string
	serverNameLength uint32
	serverName string
	timeoutMs uint32
	clientProcessFlagLength uint32
	clientProcessFlag string
	innerCode uint32
	innerMsgLength uint32
	innerMsg string
	connectionId uint32
	handlerNameLength uint32
	handlerName string

	outerReqIdLength uint32
	outerReqId       string
	version uint32

	body string
}

func (m *AgwMessage) ReqId() uint64 {
	return m.reqId
}

func (m *AgwMessage) SetReqId(reqId uint64) {
	m.reqId = reqId
}

func (m *AgwMessage) MessageType() uint8 {
	return m.messageType
}

func (m *AgwMessage) SetMessageType(messageType uint8) {
	m.messageType = messageType
}

func (m *AgwMessage) MessageDirection() uint8 {
	return m.messageDirection
}

func (m *AgwMessage) SetMessageDirection(messageDirection uint8) {
	m.messageDirection = messageDirection
}

func (m *AgwMessage) Caller() uint8 {
	return m.caller
}

func (m *AgwMessage) SetCaller(caller uint8) {
	m.caller = caller
}

func (m *AgwMessage) ClientIp() uint64 {
	return m.clientIp
}

func (m *AgwMessage) SetClientIp(clientIp uint64) {
	m.clientIp = clientIp
}

func (m *AgwMessage) ClientVpcId() string {
	return m.clientVpcId
}

func (m *AgwMessage) SetClientVpcId(clientVpcId string) {
	m.clientVpcId = clientVpcId
}

func (m *AgwMessage) ServerName() string {
	return m.serverName
}

func (m *AgwMessage) SetServerName(serverName string) {
	m.serverName = serverName
}

func (m *AgwMessage) TimeoutMs() uint32 {
	return m.timeoutMs
}

func (m *AgwMessage) SetTimeoutMs(timeoutMs uint32) {
	m.timeoutMs = timeoutMs
}

func (m *AgwMessage) ClientProcessFlag() string {
	return m.clientProcessFlag
}

func (m *AgwMessage) SetClientProcessFlag(clientProcessFlag string) {
	m.clientProcessFlag = clientProcessFlag
}

func (m *AgwMessage) InnerCode() uint32 {
	return m.innerCode
}

func (m *AgwMessage) SetInnerCode(innerCode uint32) {
	m.innerCode = innerCode
}

func (m *AgwMessage) InnerMsg() string {
	return m.innerMsg
}

func (m *AgwMessage) SetInnerMsg(innerMsg string) {
	m.innerMsg = innerMsg
}

func (m *AgwMessage) ConnectionId() uint32 {
	return m.connectionId
}

func (m *AgwMessage) SetConnectionId(connectionId uint32) {
	m.connectionId = connectionId
}

func (m *AgwMessage) HandlerName() string {
	return m.handlerName
}

func (m *AgwMessage) SetHandlerName(handlerName string) {
	m.handlerName = handlerName
}

func (m *AgwMessage) OuterReqId() string {
	return m.outerReqId
}

func (m *AgwMessage) SetOuterReqId(outerReqId string) {
	m.outerReqId = outerReqId
}

func (m *AgwMessage) Version() uint32 {
	return m.version
}

func (m *AgwMessage) SetVersion(version uint32) {
	m.version = version
}

func (m *AgwMessage) Body() string {
	return m.body
}

func (m *AgwMessage) SetBody(body string) {
	m.body = body
}

func NewAgwMessage() *AgwMessage {
	return &AgwMessage{
		innerCode: 0,
		innerMsg:  "ok",
		version:   1,
	}
}

func (m *AgwMessage) Decode(br *bufio.Reader) error {
	defer func() {
		if err := recover(); err != nil {
			logError("[AgwMessage] Decode recover err: %v, stack: %s", err, debug.Stack())
		}
	}()

	bodyLength := [4]byte{}
	for m := 0; m < len(bodyLength); {
		n, err := br.Read(bodyLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.bodyLength = uint32(binary.BigEndian.Uint32(bodyLength[:]))

	reqId := [8]byte{}
	for m := 0; m < len(reqId); {
		n, err := br.Read(reqId[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.reqId = uint64(binary.BigEndian.Uint64(reqId[:]))

	messageType, err := br.ReadByte()
	if err != nil {
		return err
	}
	m.messageType = uint8(messageType)

	messageDirection, err := br.ReadByte()
	if err != nil {
		return err
	}
	m.messageDirection = uint8(messageDirection)

	caller, err := br.ReadByte()
	if err != nil {
		return err
	}
	m.caller = uint8(caller)

	clientIp := [8]byte{}
	for m := 0; m < len(clientIp); {
		n, err := br.Read(clientIp[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.clientIp = uint64(binary.BigEndian.Uint64(clientIp[:]))

	clientVpcIdLength := [4]byte{}
	for m := 0; m < len(clientVpcIdLength); {
		n, err := br.Read(clientVpcIdLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.clientVpcIdLength = uint32(binary.BigEndian.Uint32(clientVpcIdLength[:]))

	clientVpcId := make([]byte, m.clientVpcIdLength)
	for m := 0; m < len(clientVpcId); {
		n, err := br.Read(clientVpcId[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.clientVpcId = string(clientVpcId)

	serverNameLength := [4]byte{}
	for m := 0; m < len(serverNameLength); {
		n, err := br.Read(serverNameLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.serverNameLength = uint32(binary.BigEndian.Uint32(serverNameLength[:]))

	serverName := make([]byte, m.serverNameLength)
	for m := 0; m < len(serverName); {
		n, err := br.Read(serverName[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.serverName = string(serverName)

	timeoutMs := [4]byte{}
	for m := 0; m < len(timeoutMs); {
		n, err := br.Read(timeoutMs[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.timeoutMs = uint32(binary.BigEndian.Uint32(timeoutMs[:]))

	clientProcessFlagLength := [4]byte{}
	for m := 0; m < len(clientProcessFlagLength); {
		n, err := br.Read(clientProcessFlagLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.clientProcessFlagLength = uint32(binary.BigEndian.Uint32(clientProcessFlagLength[:]))

	clientProcessFlag := make([]byte, m.clientProcessFlagLength)
	for m := 0; m < len(clientProcessFlag); {
		n, err := br.Read(clientProcessFlag[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.clientProcessFlag = string(clientProcessFlag)

	innerCode := [4]byte{}
	for m := 0; m < len(innerCode); {
		n, err := br.Read(innerCode[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.innerCode = uint32(binary.BigEndian.Uint32(innerCode[:]))

	innerMsgLength := [4]byte{}
	for m := 0; m < len(innerMsgLength); {
		n, err := br.Read(innerMsgLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.innerMsgLength = uint32(binary.BigEndian.Uint32(innerMsgLength[:]))

	innerMsg := make([]byte, m.innerMsgLength)
	for m := 0; m < len(innerMsg); {
		n, err := br.Read(innerMsg[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.innerMsg = string(innerMsg)

	connectionId := [4]byte{}
	for m := 0; m < len(connectionId); {
		n, err := br.Read(connectionId[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.connectionId = uint32(binary.BigEndian.Uint32(connectionId[:]))

	handlerNameLength := [4]byte{}
	for m := 0; m < len(handlerNameLength); {
		n, err := br.Read(handlerNameLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.handlerNameLength = uint32(binary.BigEndian.Uint32(handlerNameLength[:]))

	handlerName := make([]byte, m.handlerNameLength)
	for m := 0; m < len(handlerName); {
		n, err := br.Read(handlerName[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.handlerName = string(handlerName)

	outerReqIdLength := [4]byte{}
	for m := 0; m < len(outerReqIdLength); {
		n, err := br.Read(outerReqIdLength[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.outerReqIdLength = uint32(binary.BigEndian.Uint32(outerReqIdLength[:]))

	outerReqId := make([]byte, m.outerReqIdLength)
	for m := 0; m < len(outerReqId); {
		n, err := br.Read(outerReqId[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.outerReqId = string(outerReqId)

	version := [4]byte{}
	for m := 0; m < len(version); {
		n, err := br.Read(version[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	m.version = uint32(binary.BigEndian.Uint32(version[:]))

	body := make([]byte, m.bodyLength)
	for m := 0; m < len(body); {
		n, err := br.Read(body[m:])
		if e := handleReadResult(n, err); e != nil {
			return e
		}
		m += n
	}
	if m.version == AllCompress {
		m.body, err = tools.DecompressByGzip(body)
	} else if m.version == RequestCompress && m.messageDirection == MessageDirectionRequest {
		m.body, err = tools.DecompressByGzip(body)
	} else if m.version == ResponseCompress && m.messageDirection == MessageDirectionResponse {
		m.body, err = tools.DecompressByGzip(body)
	} else {
		m.body = string(body)
	}
	return nil
}

func handleReadResult(n int, err error) error {
	if err != nil {
		return err
	}

	if n <= 0 {
		return errors.New("read bytes less than 0")
	}

	return nil
}

func (m *AgwMessage) Encode() ([]byte, bool) {

	data := make([]byte, 0)

	var body []byte
	// 根据 version 来判断是否压缩
	var err error
	if m.version == AllCompress {
		body, err = tools.CompressByGzip(m.body)
	} else if m.version == RequestCompress && m.messageDirection == MessageDirectionRequest {
		body, err = tools.CompressByGzip(m.body)
	} else if m.version == ResponseCompress && m.messageDirection == MessageDirectionResponse {
		body, err = tools.CompressByGzip(m.body)
	} else {
		body = []byte(m.body)
	}

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(body)))

	data = append(data, make([]byte, 8)...)
	binary.BigEndian.PutUint64(data[len(data)-8:len(data)], m.reqId)

	data = append(data, byte(m.messageType))

	data = append(data, byte(m.messageDirection))

	data = append(data, byte(m.caller))

	data = append(data, make([]byte, 8)...)
	binary.BigEndian.PutUint64(data[len(data)-8:len(data)], m.clientIp)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.clientVpcId)))

	data = append(data, m.clientVpcId...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.serverName)))

	data = append(data, m.serverName...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], m.timeoutMs)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.clientProcessFlag)))

	data = append(data, m.clientProcessFlag...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], m.innerCode)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.innerMsg)))

	data = append(data, m.innerMsg...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], m.connectionId)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.handlerName)))

	data = append(data, m.handlerName...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], uint32(len(m.outerReqId)))

	data = append(data, m.outerReqId...)

	data = append(data, make([]byte, 4)...)
	binary.BigEndian.PutUint32(data[len(data)-4:len(data)], m.version)

	if err != nil {
		logger.Warnf("[AGW] Compress message err, %v", err)
		return data, false
	}
	data = append(data, body...)
	return data, true
}

func (m *AgwMessage) getSyncId() string {
	return fmt.Sprintf("%s-%d-%s-%d", m.clientVpcId, m.clientIp, m.clientProcessFlag, m.reqId)
}

func (m *AgwMessage) TimestampUtil() *timestampUtil {
	return m.tsUtil
}

func (m *AgwMessage) SetTimestampUtil(tsUtil *timestampUtil) {
	m.tsUtil = tsUtil
}
