package aliyun

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"github.com/aliyun/aliyun-ahas-go-sdk/service"
	"github.com/aliyun/aliyun-ahas-go-sdk/tools"
)

const (
	AliyuncsDomain = "aliyuncs.com"
	EcsVpcUrl      = "http://100.100.100.200/latest/meta-data/"
	CnPublic       = "cn-public"
)

var endpointMap = map[string]string{
	"pre-cn-hangzhou":     "pre.proxy.ahas.aliyun.com:9527",
	"prod-cn-hangzhou":    "proxy.ahas.cn-hangzhou.aliyuncs.com:9527",
	"prod-cn-beijing":     "proxy.ahas.cn-beijing.aliyuncs.com:9527",
	"prod-cn-shenzhen":    "proxy.ahas.cn-shenzhen.aliyuncs.com:9527",
	"prod-cn-shanghai":    "proxy.ahas.cn-shanghai.aliyuncs.com:9527",
	"prod-cn-zhangjiakou": "proxy.ahas.cn-zhangjiakou.aliyuncs.com:9527",
	"prod-cn-public":      "ahas-proxy.aliyuncs.com:8848",
}

func GetAhasProxyEndpoint(key string) (string, bool) {
	v, ok := endpointMap[key]
	return v, ok
}

var acmEndpointMap = map[string]string{
	"cn-qingdao":     "addr-qd-internal.edas.aliyun.com",
	"cn-beijing":     "addr-bj-internal.edas.aliyun.com",
	"cn-hangzhou":    "addr-hz-internal.edas.aliyun.com",
	"cn-shanghai":    "addr-sh-internal.edas.aliyun.com",
	"cn-shenzhen":    "addr-sz-internal.edas.aliyun.com",
	"cn-zhangjiakou": "addr-cn-zhangjiakou-internal.edas.aliyun.com",
	"cn-hongkong":    "addr-hk-internal.edas.aliyuncs.com",
	"ap-southeast-1": "addr-singapore-internal.edas.aliyun.com",
	"cn-public":      "acm.aliyun.com",
}

func GetAcmEndpoint(key string) (string, bool) {
	v, ok := acmEndpointMap[key]
	return v, ok
}

var tlsEndpointMap = map[string]string{
	"pre-cn-hangzhou":     "pre.proxy.ahas.tls.aliyuncs.com:9528",
	"prod-cn-hangzhou":    "proxy.ahas.cn-hangzhou.tls.aliyuncs.com:9528",
	"prod-cn-beijing":     "proxy.ahas.cn-beijing.tls.aliyuncs.com:9528",
	"prod-cn-shenzhen":    "proxy.ahas.cn-shenzhen.tls.aliyuncs.com:9528",
	"prod-cn-shanghai":    "proxy.ahas.cn-shanghai.tls.aliyuncs.com:9528",
	"prod-cn-zhangjiakou": "proxy.ahas.cn-zhangjiakou.tls.aliyuncs.com:9528",
	"prod-cn-public":      "proxy.ahas.cn-public.tls.aliyuncs.com:9528",
}

func GetAhasProxyTlsEndpoint(key string) (string, bool) {
	v, ok := tlsEndpointMap[key]
	return v, ok
}

var channel *Channel
var once sync.Once

type Channel struct {
	*service.Controller
}

type VpcEcsMetadata struct {
	IsVpc      bool
	VpcId      string
	Ip         string
	HostName   string
	RegionId   string
	InstanceId string
	Uid        string
}

// GetInstance
func GetInstance() *Channel {
	once.Do(func() {
		channel = &Channel{}
		channel.Controller = service.NewController(channel)
	})
	return channel
}

// RetrieveVpcMetadata retrieves the metadata of current ECS instance or container.
func RetrieveVpcMetadata() (*VpcEcsMetadata, error) {
	vpcEcs := &VpcEcsMetadata{}
	vpcEcs.VpcId = getVpcId()
	if vpcEcs.VpcId == "" {
		// retry
		time.Sleep(time.Second)
		vpcEcs.VpcId = getVpcId()
	}
	if vpcEcs.VpcId == "" {
		return nil, fmt.Errorf("get vpc id info failed")
	}
	vpcEcs.RegionId = getRegionId()
	if vpcEcs.RegionId == "" {
		return nil, fmt.Errorf("failed to get regionId")
	}
	vpcEcs.Ip = getPrivateIpv4()
	if vpcEcs.Ip == "" {
		return nil, fmt.Errorf("get ecs ip info failed")
	}
	vpcEcs.HostName = getHostName()
	vpcEcs.InstanceId = getInstanceId()
	if vpcEcs.InstanceId == "" {
		return nil, fmt.Errorf("get ecs id info failed")
	}
	vpcEcs.Uid = getOwnerAccountId()
	if vpcEcs.Uid == "" {
		return nil, fmt.Errorf("get vpc uid info failed")
	}
	return vpcEcs, nil
}

func (channel *Channel) DoStart() error {
	return nil
}

func (channel *Channel) DoStop() error {
	return nil
}

// Download file from AliCloud OSS.
func Download(destFileFullPath, region, originalFilePath string, isPrivate bool) error {
	if channel.Ctx.Err() != nil {
		return fmt.Errorf("aliyun channel disabled")
	}
	file, err := os.Create(destFileFullPath)

	if err != nil {
		return err
	}
	os.Chmod(destFileFullPath, 0744)
	defer file.Close()
	url := GetOssUrl(region, originalFilePath, isPrivate)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("response code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

//GetOssUrl return oss url of the file
func GetOssUrl(region, originalFilePath string, isPrivate bool) string {
	var urlFormat = ""
	if isPrivate {
		// http://ahas-cn-hangzhou.oss-cn-hangzhou-internal.aliyuncs.com/xxx
		urlFormat = "http://%s-%s.oss-%s-internal.%s/%s"
		return fmt.Sprintf(urlFormat, tools.Constant.Bucket, region, region, AliyuncsDomain, originalFilePath)
	} else {
		// http://ahasoss-cn-hangzhou.oss-cn-hangzhou.aliyuncs.com/xx
		urlFormat = "http://%s-cn-public.oss-cn-hangzhou.%s/%s"
		return fmt.Sprintf(urlFormat, tools.Constant.Bucket, AliyuncsDomain, originalFilePath)
	}
}

//getVpcId
func getVpcId() string {
	return getRemoteMessage(EcsVpcUrl + "vpc-id")
}

//getPrivateIpv4
func getPrivateIpv4() string {
	return getRemoteMessage(EcsVpcUrl + "private-ipv4")
}

//getInstanceId
func getInstanceId() string {
	return getRemoteMessage(EcsVpcUrl + "instance-id")
}

//getOwnerAccountId
func getOwnerAccountId() string {
	return getRemoteMessage(EcsVpcUrl + "owner-account-id")
}

//getHostName
func getHostName() string {
	return getRemoteMessage(EcsVpcUrl + "hostname")
}

func getRegionId() string {
	return getRemoteMessage(EcsVpcUrl + "region-id")
}

// get response message from url
func getRemoteMessage(url string) string {
	transport := http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 2*time.Second)
		},
	}
	client := http.Client{
		Transport: &transport,
	}
	resp, err := client.Get(url)
	if err != nil {
		logger.Warnf("Failed to get metadata from VPC: %s", err.Error())
		return ""
	}
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("Failed to read response when getting VPC metadata: %s", err.Error())
		return ""
	}
	result := string(bytes)
	if resp.StatusCode != 200 {
		logger.Warnf("VPC metadata: the response code is %d, message is: %s", resp.StatusCode, result)
		return ""
	}
	return result
}
