package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/sentinel-golang/core/base"
	sentinelConf "github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/log/metric"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
)

const (
	FetchMetricCommandName = "metric"
)

type FetchMetricHandler struct {
	searcher metric.MetricSearcher
}

func NewFetchMetricHandler() *FetchMetricHandler {
	s, _ := metric.NewDefaultMetricSearcher(sentinelConf.LogBaseDir(),
		metric.FormMetricFileName(sentinelConf.AppName(), sentinelConf.LogUsePid()))
	return &FetchMetricHandler{searcher: s}
}

func (h *FetchMetricHandler) Handle(request *transport.Request) *transport.Response {
	// TODO: handle panic
	var startTime, endTime uint64
	var err error
	startTimeStr := request.Params["startTime"]
	endTimeStr := request.Params["endTime"]
	maxLinesStr := request.Params["maxLines"]
	identity := request.Params["identity"]
	var list []*base.MetricItem
	if startTime, err = strconv.ParseUint(startTimeStr, 10, 64); err != nil {
		return transport.ReturnFail(transport.Code[transport.ServerError], "empty or bad startTime: "+startTimeStr)
	}
	if endTimeStr != "" {
		if endTime, err = strconv.ParseUint(endTimeStr, 10, 64); err != nil {
			return transport.ReturnFail(transport.Code[transport.ServerError], "Bad endTime: "+endTimeStr)
		}
		// Here empty resource name indicates "all".
		if list, err = h.searcher.FindByTimeAndResource(startTime, endTime, identity); err != nil {
			return transport.ReturnFail(transport.Code[transport.ServerError], fmt.Sprintf("Error when retrieving metrics: %v", err.Error()))
		}
	} else {
		var maxLines uint64 = 6000
		if maxLinesStr != "" {
			if maxLines, err = strconv.ParseUint(maxLinesStr, 10, 32); err != nil {
				return transport.ReturnFail(transport.Code[transport.ServerError], "Bad maxLines: "+maxLinesStr)
			}
		}
		if maxLines > 12000 {
			maxLines = 12000
		}
		if list, err = h.searcher.FindFromTimeWithMaxLines(startTime, uint32(maxLines)); err != nil {
			return transport.ReturnFail(transport.Code[transport.ServerError], fmt.Sprintf("Error when retrieving metrics: %v", err.Error()))
		}
	}
	if list == nil {
		list = make([]*base.MetricItem, 0)
	}
	if identity == "" {
		list = append(list, h.fetchCpuAndLoadMetric()...)
	}
	b := strings.Builder{}
	for _, item := range list {
		str, err := item.ToThinString()
		if err != nil {
			return transport.ReturnFail(transport.Code[transport.ServerError], fmt.Sprintf("Unexpected error: %v", err.Error()))
		}
		b.Write([]byte(str))
		b.WriteByte('\n')
	}
	result := b.String()
	return transport.ReturnSuccess(result)
}

func (h *FetchMetricHandler) fetchCpuAndLoadMetric() []*base.MetricItem {
	list := make([]*base.MetricItem, 0)
	t := util.CurrentTimeMillis() / 1000 * 1000
	load1 := system.CurrentLoad()
	cpuUsage := system.CurrentCpuUsage()
	if load1 > 0 {
		mi := &base.MetricItem{Resource: "__system_load__", Timestamp: t, PassQps: uint64(load1 * 10000)}
		list = append(list, mi)
	}
	if cpuUsage > 0 {
		mi := &base.MetricItem{Resource: "__cpu_usage__", Timestamp: t, PassQps: uint64(cpuUsage * 10000)}
		list = append(list, mi)
	}
	return list
}
