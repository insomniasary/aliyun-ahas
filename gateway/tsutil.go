package gateway

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

type timestampUtil struct {
	perfMap       map[string]int64
	reqId         uint64
	outerReqId    string
	lastTimestamp int64
	firstMarkDate string

	//v2
	vpcId       string
	processFlag string
	ip          string
	needMark    bool
}

func (this *timestampUtil) ReqId() uint64 {
	return this.reqId
}

func (this *timestampUtil) SetReqId(reqId uint64) {
	this.reqId = reqId
}

func (this *timestampUtil) OuterReqId() string {
	return this.outerReqId
}

func newTimestampUtil(reqId uint64, outerReqId string) *timestampUtil {
	return &timestampUtil{
		perfMap:       make(map[string]int64, 0),
		reqId:         reqId,
		outerReqId:    outerReqId,
		lastTimestamp: 0,
		firstMarkDate: "",
		needMark:      true,
	}
}

func newTimestampUtilV3() *timestampUtil {
	return &timestampUtil{
		perfMap:       make(map[string]int64, 0),
		reqId:         0,
		outerReqId:    "0",
		lastTimestamp: 0,
		firstMarkDate: "",
		needMark:      true,
	}
}

func newTimestampUtilV2(outerReqId string, vpcId string, processFlag string, ip string) *timestampUtil {
	return &timestampUtil{
		perfMap:       make(map[string]int64, 0),
		outerReqId:    outerReqId,
		lastTimestamp: 0,
		firstMarkDate: "",
		vpcId:         vpcId,
		processFlag:   processFlag,
		ip:            ip,
		needMark:      true,
	}
}

func (this *timestampUtil) NeedMark() bool {
	return this.needMark
}

func (this *timestampUtil) SetNeedMark(needMark bool) {
	this.needMark = needMark
}

func (this *timestampUtil) mark(phase string) {
	if !this.NeedMark() {
		return
	}

	if this.lastTimestamp == 0 {
		this.firstMarkDate = time.Now().String()[0:23]
		this.perfMap[phase] = 0
		this.lastTimestamp = nowMilliSecond()
		return
	}

	now := nowMilliSecond()
	this.perfMap[phase] = now - this.lastTimestamp
	this.lastTimestamp = now
}

func (this *timestampUtil) GetResult() string {
	if !this.NeedMark() {
		return "do not need mark"
	}

	var buf bytes.Buffer

	buf.WriteString("time statistics [reqId:")
	buf.WriteString(strconv.FormatUint(this.reqId, 10))
	buf.WriteString(", outerReqId:")
	buf.WriteString(this.outerReqId)
	buf.WriteString(", firstMark:")
	buf.WriteString(this.firstMarkDate)
	buf.WriteString(", info:")
	buf.WriteString(fmt.Sprint(this.perfMap))
	buf.WriteString("]")

	return buf.String()
}

func (this *timestampUtil) GetResultV2() string {
	if !this.NeedMark() {
		return "do not need mark"
	}

	var buf bytes.Buffer

	buf.WriteString("time statistics [reqId:")
	buf.WriteString(strconv.FormatUint(this.reqId, 10))
	buf.WriteString(", outerReqId:")
	buf.WriteString(this.outerReqId)

	buf.WriteString(", vpcId:")
	buf.WriteString(this.vpcId)

	buf.WriteString(", processFlag:")
	buf.WriteString(this.processFlag)

	buf.WriteString(", ip:")
	buf.WriteString(this.ip)

	buf.WriteString(", firstMark:")
	buf.WriteString(this.firstMarkDate)
	buf.WriteString(", info:")
	buf.WriteString(fmt.Sprint(this.perfMap))
	buf.WriteString("]")

	return buf.String()
}

func (this *timestampUtil) GetResultV3() string {
	if !this.NeedMark() {
		return "do not need mark"
	}

	var buf bytes.Buffer

	buf.WriteString("firstMark:")
	buf.WriteString(this.firstMarkDate)
	buf.WriteString(", info:")
	buf.WriteString(fmt.Sprint(this.perfMap))
	buf.WriteString("]")

	return buf.String()
}

func nowMilliSecond() int64 {
	return time.Now().UnixNano() / 1000000
}
