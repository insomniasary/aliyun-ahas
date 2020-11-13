package gateway

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	connectTimeoutSec = 5
)

var createOnce sync.Once
var poolInstance *ConnectionPool

type AgwConn struct {
	connId   uint32
	conn     *net.Conn
	pool     *ConnectionPool
	channels sync.Map
}

func (c *AgwConn) writeSync(msg *AgwMessage) (*AgwMessage, error) {

	msgBytes, ok := msg.Encode()

	if !ok {
		return nil, errors.New("encode wrong")
	}

	if _, e := (*c.conn).Write(msgBytes); e != nil {
		return nil, e
	}

	response, err := wait(c, msg)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *AgwConn) write(msg *AgwMessage) error {
	msgBytes, ok := msg.Encode()
	if !ok {
		return errors.New("encode wrong")
	}

	if _, e := (*c.conn).Write(msgBytes); e != nil {
		logWarnf("[AGW] gateway write err: %+v", e.Error())
		return e
	}

	return nil
}

func (c *AgwConn) close() {
	if c == nil {
		return
	}
	logInfof("[AGW] Close connection, connId : %d", c.connId)

	c.pool.remove(c.connId)
	(*c.conn).Close()

	connClosedMsg := NewAgwMessage()
	connClosedMsg.SetInnerMsg(ErrorMsgConnClosed)

	c.channels.Range(func(k, v interface{}) bool {
		if channel, ok := v.(chan *AgwMessage); ok {
			channel <- connClosedMsg
		}
		return true
	})
}

type ConnectionPool struct {
	ring *ring
	pool sync.Map
	lock sync.Mutex
	size uint32
}

func getConnectionPoolInstance(size uint32) *ConnectionPool {

	if size <= 0 {
		return nil
	}

	if poolInstance != nil {
		return poolInstance
	}
	createOnce.Do(func() {
		if poolInstance == nil {
			poolInstance = new(ConnectionPool)
			poolInstance.ring = newRing()
			poolInstance.size = size
			for i := uint32(0); i < size; i++ {
				poolInstance.ring.add(i)
			}
		}
	})
	return poolInstance
}

func (p *ConnectionPool) get() (*AgwConn, error) {
	var connId uint32
	if value, ok := p.ring.next().(uint32); ok {
		connId = value
	} else {
		return nil, errors.New("connId should be uint32")
	}

	if conn, ok := p.pool.Load(connId); ok {
		if value, ok := conn.(*AgwConn); ok {
			return value, nil
		} else {
			return nil, errors.New("connection type should be *AgwConn")
		}
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if conn, ok := p.pool.Load(connId); ok {
		if value, ok := conn.(*AgwConn); ok {
			return value, nil
		} else {
			return nil, errors.New("unknown connection type")
		}
	}

	gatewayIp := GetAgwClientInstance().config.GatewayIp
	gatewayPort := GetAgwClientInstance().config.GatewayPort
	var conn net.Conn
	var err error
	// tls conn or not
	if GetAgwClientInstance().config.TlsFlag {
		conn, err = getTlsConn(gatewayIp, gatewayPort)
		// retry once
		if err != nil {
			logger.Warnf("[AGW] Get TLS connection err, %v, retry again", err)
			err := checkOrDownloadCert()
			if err != nil {
				return nil, err
			}
			conn, err = getTlsConn(gatewayIp, gatewayPort)
		}
	} else {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", gatewayIp, gatewayPort), connectTimeoutSec*time.Second)
	}
	if err != nil {
		return nil, err
	}
	logInfof("AGW connect [%s:%d] success, connectionId: %d", gatewayIp, gatewayPort, connId)

	agwConn := &AgwConn{
		connId: connId,
		conn:   &conn,
		pool:   p,
	}

	p.pool.Store(connId, agwConn)

	go runReaderCoroutine(agwConn)

	return agwConn, nil
}

func getTlsConn(gatewayIp string, gatewayPort uint32) (net.Conn, error) {
	certFile, err := os.OpenFile(CertPath, os.O_RDONLY, 0664)
	if err != nil {
		return nil, fmt.Errorf("open cert file failed, %v", err)
	}
	certBytes, err := ioutil.ReadAll(certFile)
	if err != nil {
		return nil, fmt.Errorf("read cert file failed, %v", err)
	}
	certPool := x509.NewCertPool()
	ok := certPool.AppendCertsFromPEM(certBytes)
	if !ok {
		return nil, fmt.Errorf("parse cert file failed")
	}
	conf := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            certPool,
	}
	dialer := &net.Dialer{Timeout: connectTimeoutSec * time.Second}
	return tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", gatewayIp, gatewayPort), conf)
}

func (p *ConnectionPool) remove(connId uint32) {
	p.pool.Delete(connId)
}

func StringIpToUint64(ip string) uint64 {
	ipSegs := strings.Split(ip, ".")
	var ipUint64 uint64 = 0
	var pos uint = 24
	var tmp uint64 = 0
	for _, ipSeg := range ipSegs {
		tempInt, _ := strconv.Atoi(ipSeg)
		tmp = uint64(tempInt)
		tmp = tmp << pos
		ipUint64 = ipUint64 | tmp
		pos -= 8
	}
	return ipUint64
}
