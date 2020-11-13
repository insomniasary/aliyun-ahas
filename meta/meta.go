package meta

import (
	"net"
	"os"
	"strconv"

	"github.com/aliyun/aliyun-ahas-go-sdk/aliyun"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/pkg/errors"
)

const (
	CurrentSdkVersion = "1.0.3"

	GoSDK = "GO_SDK"
)
const (
	Host = iota
	Container
)

var metadata = &Meta{
	version: CurrentSdkVersion,
	tidChan: make(chan string, 5),
}

func InitMetadata(license, namespace, env string, secureTransport bool) (*Meta, error) {
	metadata.license = license
	metadata.namespace = namespace
	metadata.deployEnv = env

	if license == "" {
		vpcEcs, err := aliyun.RetrieveVpcMetadata()
		if err != nil || vpcEcs.Uid == "" {
			return nil, errors.New("cannot find AHAS license")
		}
		metadata.regionId = vpcEcs.RegionId
		metadata.inVpc = true
		metadata.vpcId = vpcEcs.VpcId
		metadata.ip = vpcEcs.Ip
		metadata.hostName = vpcEcs.HostName
		metadata.pid = resolveProcessId()
		metadata.instanceId = vpcEcs.InstanceId
		metadata.uid = vpcEcs.Uid
	} else {
		metadata.regionId = aliyun.CnPublic
		metadata.inVpc = false
		metadata.vpcId = license
		ip, err := resolvePrivateIp()
		if err != nil {
			return nil, errors.Wrap(err, "cannot resolve private IP")
		}
		metadata.ip = ip
		metadata.hostName = resolveHostName()
		metadata.pid = resolveProcessId()
		metadata.instanceId = resolveHostName()
		// Get UID from server side
		metadata.uid = ""
	}

	envKey := env + "-" + metadata.regionId
	var endpoint string
	var envSupported bool
	if secureTransport {
		endpoint, envSupported = aliyun.GetAhasProxyTlsEndpoint(envKey)
	} else {
		endpoint, envSupported = aliyun.GetAhasProxyEndpoint(envKey)
	}

	if !envSupported || endpoint == "" {
		logger.Warn("No available AHAS endpoint, env not supported: " + envKey)
		return nil, errors.New("No available AHAS endpoint, env not supported: " + envKey)
	}
	metadata.ahasEndpoint = endpoint
	metadata.version = CurrentSdkVersion

	return metadata, nil
}

func License() string {
	return metadata.license
}

func Namespace() string {
	return metadata.namespace
}

func DeployEnv() string {
	return metadata.deployEnv
}

func IsPrivate() bool {
	return metadata.inVpc
}

func CurrentVersion() string {
	return metadata.version
}

func RegionId() string {
	return metadata.regionId
}

func VpcId() string {
	return metadata.vpcId
}

func Pid() string {
	return metadata.pid
}

func Cid() string {
	return metadata.cid
}

func LocalIp() string {
	return metadata.ip
}

func DebugEnabled() bool {
	return metadata.debugging
}

func resolveProcessId() string {
	return strconv.Itoa(os.Getpid())
}

func resolveHostName() string {
	name, err := os.Hostname()
	if err != nil {
		logger.Warnf("Failed to get hostname: %v", err)
		return ""
	}
	return name
}

func resolvePrivateIp() (string, error) {
	ifs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, i := range ifs {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			logger.Warnf("Failed to list addresses of the interface <%s>: %v", i.Name, err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Cannot get host ip address")
}
