package handler

import (
	"encoding/json"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
)

const (
	GetResourceNodeCommandName = "clusterNode"
)

type NodeVO struct {
	Id          string `json:"id,omitempty"`
	ParentId    string `json:"parentId,omitempty"`
	Resource    string `json:"resource"`
	Concurrency int32  `json:"threadNum"`
	PassQps     uint64 `json:"passQps"`
	BlockQps    uint64 `json:"blockQps"`
	TotalQps    uint64 `json:"totalQps"`
	AvgRt       uint64 `json:"averageRt"`
	CompleteQps uint64 `json:"successQps"`
	Timestamp   uint64 `json:"timestamp"`
}

func NodeVoFromReal(n *stat.ResourceNode) *NodeVO {
	pass := uint64(n.GetQPS(base.MetricEventPass))
	block := uint64(n.GetQPS(base.MetricEventBlock))
	return &NodeVO{
		Resource:    n.ResourceName(),
		Concurrency: n.CurrentGoroutineNum(),
		PassQps:     pass,
		BlockQps:    block,
		TotalQps:    pass + block,
		AvgRt:       uint64(n.AvgRT()),
		CompleteQps: uint64(n.GetQPS(base.MetricEventComplete)),
		Timestamp:   util.CurrentTimeMillis(),
	}
}

type ResourceNodeHandler struct {
}

func (r *ResourceNodeHandler) Handle(_ *transport.Request) *transport.Response {
	nodes := stat.ResourceNodeList()
	voList := make([]*NodeVO, 0)
	for _, n := range nodes {
		voList = append(voList, NodeVoFromReal(n))
	}
	bs, err := json.Marshal(voList)
	if err != nil {
		return transport.ReturnFail(transport.Code[transport.ServerError], "bad data")
	}
	return transport.ReturnSuccess(string(bs))
}
