package config

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/aliyun/aliyun-ahas-go-sdk/heartbeat"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/sentinel/datasource"
	"github.com/aliyun/aliyun-ahas-go-sdk/transport"
	"gopkg.in/yaml.v2"
)

const (
	DeployEnvProd = "prod"
	DeployEnvPre  = "pre"
	DeployEnvTest = "test"

	DefaultNamespace = "default"

	LicenseEnvKey     = "AHAS_LICENSE"
	NamespaceEnvKey   = "AHAS_NAMESPACE"
	EnvironmentEnvKey = "AHAS_ENV"

	ConfFileEnvKey = "AHAS_CONFIG_FILE_PATH"
)

type Config struct {
	License    string            `yaml:"license"`
	Namespace  string            `yaml:"namespace"`
	Env        string            `yaml:"env"`
	Transport  transport.Config  `yaml:"transport"`
	Heartbeat  heartbeat.Config  `yaml:"heartbeat"`
	DataSource datasource.Config `yaml:"datasource"`
}

func NewDefaultConfig() *Config {
	return &Config{
		Namespace: DefaultNamespace,
		Env:       DeployEnvProd,
		Transport: transport.Config{
			TimeoutMs: 3000,
			Secure:    true,
		},
		Heartbeat: heartbeat.Config{
			PeriodMs: 5000,
		},
		DataSource: datasource.Config{
			TimeoutMs:        datasource.DefaultTimeoutMs,
			ListenIntervalMs: datasource.DefaultListenIntervalMs,
		},
	}
}

var localConf = NewDefaultConfig()

func InitConfig() error {
	return InitConfigFromFile("")
}

func resolveConfigFilePath(filePath string) string {
	if !util.IsBlank(filePath) {
		return filePath
	}
	// If the file path is absent, Sentinel will try to resolve it from the system env.
	if filePath = os.Getenv(ConfFileEnvKey); !util.IsBlank(filePath) {
		return filePath
	}
	if filePath = os.Getenv(config.ConfFilePathEnvKey); !util.IsBlank(filePath) {
		return filePath
	}
	return config.DefaultConfigFilename
}

func InitConfigFromFile(p string) error {
	filePath := resolveConfigFilePath(p)
	err := loadConfFromYamlFile(filePath)
	if err != nil {
		return err
	}

	loadConfFromSystemEnv()
	if err = checkAndFillDefaultValues(); err != nil {
		return err
	}

	return nil
}

func checkAndFillDefaultValues() error {
	if localConf.DataSource.TimeoutMs == 0 {
		localConf.DataSource.TimeoutMs = datasource.DefaultTimeoutMs
	}
	if localConf.DataSource.ListenIntervalMs == 0 {
		localConf.DataSource.ListenIntervalMs = datasource.DefaultListenIntervalMs
	}
	if localConf.DataSource.ListenIntervalMs < localConf.DataSource.TimeoutMs {
		return errors.New("DataSource.ListenIntervalMs should be greater than DataSource.TimeoutMs")
	}
	return nil
}

func loadConfFromYamlFile(filePath string) error {
	if filePath == config.DefaultConfigFilename {
		if _, err := os.Stat(filePath); err != nil {
			return nil
		}
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	data := &struct {
		Version string
		AHAS    *Config `yaml:"ahas"`
	}{
		AHAS: localConf,
	}
	err = yaml.Unmarshal(content, &data)
	if err != nil {
		return err
	}
	logger.Infof("Resolving AHAS config from: %s", filePath)
	return nil
}

func loadConfFromSystemEnv() {
	if license := os.Getenv(LicenseEnvKey); !util.IsBlank(license) {
		localConf.License = license
	}
	if namespace := os.Getenv(NamespaceEnvKey); !util.IsBlank(namespace) {
		localConf.Namespace = namespace
	}
	if ahasEnv := os.Getenv(EnvironmentEnvKey); !util.IsBlank(ahasEnv) {
		localConf.Env = ahasEnv
	}
}

func License() string {
	return localConf.License
}

func Namespace() string {
	return localConf.Namespace
}

func DeployEnv() string {
	return localConf.Env
}

func TransportConfig() transport.Config {
	return localConf.Transport
}

func HeartbeatConfig() heartbeat.Config {
	return localConf.Heartbeat
}

func DataSourceConfig() datasource.Config {
	return localConf.DataSource
}
