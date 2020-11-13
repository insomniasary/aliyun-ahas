package datasource

import (
	"encoding/json"
	"time"

	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	sentinelConf "github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/hotspot"
	"github.com/alibaba/sentinel-golang/core/system"
	sentinelLogger "github.com/alibaba/sentinel-golang/logging"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
)

const (
	AcmGroupId = "ahas-sentinel"

	FlowRuleDataIdPrefix            = "flow-rule-"
	SystemRuleDataIdPrefix          = "system-rule-"
	CircuitBreakingRuleDataIdPrefix = "degrade-rule-"
	ParamFlowRuleDataIdPrefix       = "param-flow-rule-"
)

func formFlowRuleDataId(userId, namespace, appName string) string {
	return FlowRuleDataIdPrefix + userId + "-" + namespace + "-" + appName
}

func formSystemRuleDataId(userId, namespace, appName string) string {
	return SystemRuleDataIdPrefix + userId + "-" + namespace + "-" + appName
}

func formCircuitBreakingRuleDataId(userId, namespace, appName string) string {
	return CircuitBreakingRuleDataIdPrefix + userId + "-" + namespace + "-" + appName
}

func formParamFlowRuleDataId(userId, namespace, appName string) string {
	return ParamFlowRuleDataIdPrefix + userId + "-" + namespace + "-" + appName
}

func InitAcm(acmHost string, conf Config, m *meta.Meta) error {
	ch := m.TidChan()
	select {
	case <-ch:
		break
	case <-time.After(30 * time.Second):
		return errors.New("wait AHAS transport timeout")
	}

	clientConfig := constant.ClientConfig{
		TimeoutMs:      conf.TimeoutMs,
		ListenInterval: conf.ListenIntervalMs,
		NamespaceId:    m.Tid(),
		Endpoint:       acmHost + ":8080",
	}
	configClient, err := clients.CreateConfigClient(map[string]interface{}{
		"clientConfig": clientConfig,
	})
	if err != nil {
		return err
	}

	flowRuleDataId := formFlowRuleDataId(m.Uid(), meta.Namespace(), sentinelConf.AppName())
	err = configClient.ListenConfig(vo.ConfigParam{
		Group:  AcmGroupId,
		DataId: flowRuleDataId,
		OnChange: func(namespace, group, dataId, data string) {
			onFlowRuleChange(data)
		},
	})
	if err != nil {
		return err
	}
	// Add system rule config listener.
	err = configClient.ListenConfig(vo.ConfigParam{
		Group:  AcmGroupId,
		DataId: formSystemRuleDataId(m.Uid(), meta.Namespace(), sentinelConf.AppName()),
		OnChange: func(namespace, group, dataId, data string) {
			onSystemRuleChange(data)
		},
	})
	if err != nil {
		return err
	}
	// Add circuit breaking rule config listener.
	err = configClient.ListenConfig(vo.ConfigParam{
		Group:  AcmGroupId,
		DataId: formCircuitBreakingRuleDataId(m.Uid(), meta.Namespace(), sentinelConf.AppName()),
		OnChange: func(namespace, group, dataId, data string) {
			onCircuitBreakingRuleChange(data)
		},
	})
	if err != nil {
		return err
	}
	// Add param flow rule config listener.
	err = configClient.ListenConfig(vo.ConfigParam{
		Group:  AcmGroupId,
		DataId: formParamFlowRuleDataId(m.Uid(), meta.Namespace(), sentinelConf.AppName()),
		OnChange: func(namespace, group, dataId, data string) {
			onParamFlowRuleChange(data)
		},
	})
	if err != nil {
		return err
	}

	sentinelLogger.Info("ACM data source initialized successfully")
	logger.Infof("ACM data source initialized successfully, flow dataId: %s", flowRuleDataId)
	return nil
}

func onFlowRuleChange(data string) {
	sentinelLogger.Infof("ACM data received for flow rules: %v", data)
	d := &struct {
		Version string
		Data    []*LegacyFlowRule
	}{}
	err := json.Unmarshal([]byte(data), d)
	if err != nil {
		sentinelLogger.Errorf("Failed to parse flow rules: %+v", err)
		return
	}
	arr := make([]*flow.FlowRule, 0)
	for _, r := range d.Data {
		if rule := r.ToGoRule(); rule != nil {
			arr = append(arr, rule)
		}
	}
	_, err = flow.LoadRules(arr)
	if err != nil {
		sentinelLogger.Errorf("Failed to load flow rules: %+v", err)
		return
	}
}

func onSystemRuleChange(data string) {
	sentinelLogger.Infof("ACM data received for system rules: %v", data)
	d := &struct {
		Version string
		Data    []*LegacySystemRule
	}{}
	err := json.Unmarshal([]byte(data), d)
	if err != nil {
		sentinelLogger.Errorf("Failed to parse system rules: %+v", err)
		return
	}
	arr := make([]*system.SystemRule, 0)
	for _, r := range d.Data {
		if rule := r.ToGoRule(); rule != nil {
			arr = append(arr, rule)
		}
	}
	_, err = system.LoadRules(arr)
	if err != nil {
		sentinelLogger.Errorf("Failed to load system rules: %+v", err)
		return
	}
}

func onCircuitBreakingRuleChange(data string) {
	sentinelLogger.Infof("ACM data received for circuit breaking rules: %v", data)
	d := &struct {
		Version string
		Data    []*LegacyDegradeRule
	}{}
	err := json.Unmarshal([]byte(data), d)
	if err != nil {
		sentinelLogger.Errorf("Failed to parse legacy degrade rules: %+v", err)
		return
	}
	arr := make([]*circuitbreaker.Rule, 0)
	for _, r := range d.Data {
		if rule := r.ToGoRule(); rule != nil {
			arr = append(arr, rule)
		}
	}
	_, err = circuitbreaker.LoadRules(arr)
	if err != nil {
		sentinelLogger.Errorf("Failed to load flow rules: %+v", err)
		return
	}
}

func onParamFlowRuleChange(data string) {
	sentinelLogger.Infof("ACM data received for hot-spot parameter flow rules: %v", data)
	d := &struct {
		Version string
		Data    []*LegacyParamFlowRule
	}{}
	err := json.Unmarshal([]byte(data), d)
	if err != nil {
		sentinelLogger.Errorf("Failed to parse legacy param flow rules: %+v", err)
		return
	}
	arr := make([]*hotspot.Rule, 0)
	for _, r := range d.Data {
		if rule := r.ToGoRule(); rule != nil {
			arr = append(arr, rule)
		}
	}
	_, err = hotspot.LoadRules(arr)
	if err != nil {
		sentinelLogger.Errorf("Failed to load hot-spot parameter flow rules: %+v", err)
		return
	}
}
