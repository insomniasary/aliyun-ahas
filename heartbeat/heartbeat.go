package heartbeat

import (
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
	"time"
)

const (
	DefaultPeriodMs uint64 = 5000
)

type Config struct {
	PeriodMs uint64 `yaml:"period"`
}

type heartbeat struct {
	period time.Duration
	*transport.Transport
}

// New heartbeat
func New(config Config, trans *transport.Transport) *heartbeat {
	if config.PeriodMs == 0 {
		config.PeriodMs = DefaultPeriodMs
	}
	handler := &GetPingHandler().AgwRequestHandler
	trans.RegisterHandler(transport.Ping, handler)
	return &heartbeat{
		period:    time.Duration(config.PeriodMs) * time.Millisecond,
		Transport: trans,
	}
}

//Start heartbeat service
func (beat *heartbeat) Start() *heartbeat {
	ticker := time.NewTicker(beat.period)
	go func() {
		defer tools.PrintPanicStack()
		for range ticker.C {
			uri := transport.NewUri(transport.Topology, transport.Heartbeat)
			request := transport.NewRequest()
			beat.sendHeartbeat(uri, request)
		}
	}()
	logger.Infof("AGW heartbeat service started successfully, cid: %s, ver: %s, vpcId: %s",
		meta.Cid(), meta.CurrentVersion(), meta.VpcId())
	return nil
}

// sendHeartbeat
func (beat *heartbeat) sendHeartbeat(uri transport.Uri, request *transport.Request) {
	response, err := beat.Invoke(uri, request)
	if err != nil {
		logger.Warnf("Send heartbeat failed: %s", err.Error())
		beat.record(false)
		return
	}
	if !response.Success {
		logger.Errorf("AGW heartbeat bad response: %+v", response)
		beat.record(false)
		return
	}
	beat.record(true)
}

func (beat *heartbeat) record(success bool) {
	// TODO: record snapshot of heartbeat result.
}

type HBSnapshot struct {
	Timestamp int64
	Success   bool
}
