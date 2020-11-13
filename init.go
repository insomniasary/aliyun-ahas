package ahas

import (
	"fmt"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/aliyun/aliyun-ahas-go-sdk/aliyun"
	"github.com/aliyun/aliyun-ahas-go-sdk/config"
	"github.com/aliyun/aliyun-ahas-go-sdk/heartbeat"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/meta"
	"github.com/aliyun/aliyun-ahas-go-sdk/sentinel/datasource"
	"github.com/aliyun/aliyun-ahas-go-sdk/sentinel/handler"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
	"github.com/pkg/errors"
)

func InitAhasDefault() error {
	return InitAhasFromFile("")
}

func InitAhasFromFile(filename string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	if err = sentinel.InitWithConfigFile(filename); err != nil {
		return err
	}
	if err = logger.InitLoggerDefault(); err != nil {
		return err
	}
	if err = config.InitConfigFromFile(filename); err != nil {
		return err
	}
	var m *meta.Meta
	m, err = meta.InitMetadata(config.License(), config.Namespace(),
		config.DeployEnv(), config.TransportConfig().Secure)
	if err != nil {
		return err
	}

	aliyunChannel := aliyun.GetInstance()
	if err = aliyunChannel.Start(); err != nil {
		return err
	}

	tools.InitConstant(config.DeployEnv(), m.RegionId())

	acmHost, ok := aliyun.GetAcmEndpoint(m.RegionId())
	if !ok {
		return errors.New("no available ACM endpoint for region: " + m.RegionId())
	}

	// Initialize AHAS transport module.
	tc := config.TransportConfig()
	var tsp *transport.Transport
	if tsp, err = transport.New(&tc, m); err != nil {
		return err
	}
	if tsp, err = tsp.Start(); err != nil {
		return err
	}
	registerTransportHandlers(tsp)
	// Initialize heartbeat task.
	heartbeat.New(config.HeartbeatConfig(), tsp).Start()

	go initializeAcmDataSource(acmHost, m)

	return nil
}

func initializeAcmDataSource(acmHost string, m *meta.Meta) {
	defer tools.PrintPanicStackV2("failed to init ACM data-source")
	err := datasource.InitAcm(acmHost, config.DataSourceConfig(), m)
	if err != nil {
		logging.Errorf("Failed to initialize ACM data source: %+v")
	}
}

func registerTransportHandlers(tsp *transport.Transport) {
	cnHandler := transport.NewCommonHandler(&handler.ResourceNodeHandler{})
	tsp.RegisterHandler(handler.GetResourceNodeCommandName, &cnHandler)
	metricHandler := transport.NewCommonHandler(handler.NewFetchMetricHandler())
	tsp.RegisterHandler(handler.FetchMetricCommandName, &metricHandler)
}
