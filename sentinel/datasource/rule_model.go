package datasource

import (
	"strconv"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
)

type LegacyFlowRule struct {
	// ID represents the unique ID of the rule (optional).
	ID uint64 `json:"id,omitempty"`

	// Resource represents the resource name.
	Resource string `json:"resource"`
	// LimitOrigin represents the target origin (reserved field).
	LimitOrigin string          `json:"limitApp"`
	MetricType  flow.MetricType `json:"grade"`
	// Count represents the threshold.
	Count           float64               `json:"count"`
	Strategy        flow.RelationStrategy `json:"strategy"`
	ControlBehavior flow.ControlBehavior  `json:"controlBehavior"`

	RefResource       string `json:"refResource,omitempty"`
	WarmUpPeriodSec   uint32 `json:"warmUpPeriodSec"`
	MaxQueueingTimeMs uint32 `json:"maxQueueingTimeMs"`

	// ClusterMode indicates whether the rule is for cluster flow control or local.
	ClusterMode bool `json:"clusterMode"`
}

func (lr *LegacyFlowRule) ToGoRule() *flow.FlowRule {
	return &flow.FlowRule{
		ID:                lr.ID,
		Resource:          lr.Resource,
		LimitOrigin:       lr.LimitOrigin,
		MetricType:        lr.MetricType,
		Count:             lr.Count,
		RelationStrategy:  lr.Strategy,
		ControlBehavior:   lr.ControlBehavior,
		RefResource:       lr.RefResource,
		WarmUpPeriodSec:   lr.WarmUpPeriodSec,
		MaxQueueingTimeMs: lr.MaxQueueingTimeMs,
		ClusterMode:       lr.ClusterMode,
	}
}

type LegacySystemRule struct {
	ID                uint64  `json:"id,omitempty"`
	Resource          string  `json:"resource"`
	HighestSystemLoad float64 `json:"highestSystemLoad,omitempty"`
	HighestCpuUsage   float64 `json:"highestCpuUsage,omitempty"`
	InboundQps        float64 `json:"qps,omitempty"`
	AvgRt             int64   `json:"avgRt,omitempty"`
	MaxConcurrency    int64   `json:"maxThread,omitempty"`
}

func (lr *LegacySystemRule) resolveTypeAndCount() (system.MetricType, float64) {
	if lr.AvgRt >= 0 {
		return system.AvgRT, float64(lr.AvgRt)
	}
	if lr.MaxConcurrency >= 0 {
		return system.Concurrency, float64(lr.MaxConcurrency)
	}
	if lr.InboundQps >= 0 {
		return system.InboundQPS, lr.InboundQps
	}
	if lr.HighestCpuUsage >= 0 {
		return system.CpuUsage, lr.HighestCpuUsage
	}
	if lr.HighestSystemLoad >= 0 {
		return system.Load, lr.HighestSystemLoad
	}
	return system.MetricType(404), -1
}

func (lr *LegacySystemRule) ToGoRule() *system.SystemRule {
	mt, count := lr.resolveTypeAndCount()
	adaptive := system.NoAdaptive
	if mt == system.Load || mt == system.CpuUsage {
		adaptive = system.BBR
	}
	return &system.SystemRule{
		ID:           lr.ID,
		TriggerCount: count,
		MetricType:   mt,
		Strategy:     adaptive,
	}
}

type LegacyDegradeRule struct {
	ID                 uint64  `json:"id,omitempty"`
	Resource           string  `json:"resource"`
	Threshold          float64 `json:"count"`
	Strategy           uint32  `json:"grade"`
	RetryTimeoutSec    uint32  `json:"timeWindow"`
	MinRequestAmount   uint64  `json:"minRequestAmount"`
	SlowRatioThreshold float64 `json:"slowRatioThreshold"`
	StatIntervalMs     uint32  `json:"statIntervalMs"`
}

func (lr *LegacyDegradeRule) ToGoRule() *circuitbreaker.Rule {
	rule := &circuitbreaker.Rule{
		Id:               strconv.FormatUint(lr.ID, 10),
		Resource:         lr.Resource,
		StatIntervalMs:   lr.StatIntervalMs,
		RetryTimeoutMs:   lr.RetryTimeoutSec * 1000,
		Threshold:        lr.Threshold,
		MinRequestAmount: lr.MinRequestAmount,
	}
	switch lr.Strategy {
	case 0:
		// Legacy convention: threshold is RT upper bound, and the slow ratio is an independent field
		rule.Strategy = circuitbreaker.SlowRequestRatio
		rule.Threshold = lr.SlowRatioThreshold
		rule.MaxAllowedRtMs = uint64(lr.Threshold)
		break
	case 1:
		rule.Strategy = circuitbreaker.ErrorRatio
		break
	case 2:
		rule.Strategy = circuitbreaker.ErrorCount
		break
	default:
		return nil
	}
	return rule
}

type LegacyParamFlowItem struct {
	Value     string  `json:"object"`
	Threshold float64 `json:"count"`
	ParamType string  `json:"classType"`
}

type LegacyParamFlowRule struct {
	Id         uint64             `json:"id,omitempty"`
	Resource   string             `json:"resource"`
	MetricType hotspot.MetricType `json:"grade"`
	Threshold  float64            `json:"count"`
	// ParamIndex is the index in context arguments slice.
	ParamIndex        int32                  `json:"paramIdx"`
	DurationInSec     int64                  `json:"durationInSec"`
	ControlBehavior   uint32                 `json:"controlBehavior"`
	MaxQueueingTimeMs int64                  `json:"maxQueueingTimeMs"`
	BurstCount        int64                  `json:"burstCount"`
	SpecificItems     []*LegacyParamFlowItem `json:"paramFlowItemList,omitempty"`
	// ClusterMode indicates whether the rule is for cluster flow control or local.
	ClusterMode bool `json:"clusterMode"`
}

func (lr *LegacyParamFlowRule) ToGoRule() *hotspot.Rule {
	cb := hotspot.Reject
	if lr.ControlBehavior == 2 {
		cb = hotspot.Throttling
	}
	items := make([]hotspot.SpecificValue, 0)
	//items := make(map[hotspot.SpecificValue]int64, 0)
	if len(lr.SpecificItems) > 0 {
		// Re-construct param specific items
		for _, v := range lr.SpecificItems {
			if len(v.Value) == 0 {
				continue
			}
			if v.ParamType == "int" || v.ParamType == "long" {
				items = append(items, hotspot.SpecificValue{ValKind: hotspot.KindInt, ValStr: v.Value, Threshold: int64(v.Threshold)})
			} else if v.ParamType == "bool" || v.ParamType == "boolean" {
				items = append(items, hotspot.SpecificValue{ValKind: hotspot.KindBool, ValStr: v.Value, Threshold: int64(v.Threshold)})
			} else if v.ParamType == "double" || v.ParamType == "float" {
				items = append(items, hotspot.SpecificValue{ValKind: hotspot.KindFloat64, ValStr: v.Value, Threshold: int64(v.Threshold)})
			} else {
				items = append(items, hotspot.SpecificValue{ValKind: hotspot.KindString, ValStr: v.Value, Threshold: int64(v.Threshold)})
			}
		}
	}

	return &hotspot.Rule{
		ID:                strconv.Itoa(int(lr.Id)),
		Resource:          lr.Resource,
		MetricType:        lr.MetricType,
		Threshold:         lr.Threshold,
		ControlBehavior:   cb,
		ParamIndex:        int(lr.ParamIndex),
		MaxQueueingTimeMs: lr.MaxQueueingTimeMs,
		BurstCount:        lr.BurstCount,
		DurationInSec:     lr.DurationInSec,
		ParamsMaxCapacity: 500,
		SpecificItems:     items,
	}
}
