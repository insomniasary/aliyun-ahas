package tools

import (
	"os/user"
)

const (
	Pre  = "pre"
	Test = "test"
	Prod = "prod"
)

var Repositories = map[string]*Constants{
	Pre: {
		Env:               Pre,
		RepositoryName:    "ahas",
		OSAgentRemotePath: "agent/pre",
		Bucket:            "ahas",
	},
	Prod: {
		Env:               Prod,
		RepositoryName:    "ahascr",
		OSAgentRemotePath: "agent/prod",
		Bucket:            "ahasoss",
	},
	Test: {
		Env:               Test,
		RepositoryName:    "ahas",
		OSAgentRemotePath: "agent/test",
		Bucket:            "ahas",
	},
}

var Constant *Constants

type Constants struct {
	Env               string
	RepositoryName    string
	RepositoryDomain  string
	OSAgentRemotePath string
	Bucket            string
}

func InitConstant(environment, regionId string) {
	Constant = Repositories[environment]
	if Constant == nil {
		Constant = Repositories[Prod]
	}
	if IsPublicEnv(regionId) {
		Constant.RepositoryDomain = "registry.cn-hangzhou.aliyuncs.com"
	} else {
		Constant.RepositoryDomain = "registry-vpc." + regionId + ".aliyuncs.com"
	}
}

// GetUserHome return user home.
func GetUserHome() string {
	user, err := user.Current()
	if err == nil {
		return user.HomeDir
	}
	return "/root"
}

func IsPublicEnv(regionId string) bool {
	return "cn-public" == regionId
}
