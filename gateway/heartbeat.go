package gateway

import (
	"time"
)

const (
	HeartbeatHandlerName = "HB"
	HeartbeatServerName  = "HBGATEWAY"
	HeartbeatMessageBody = "HBMSGBODY"
	HeartbeatTimeoutMs   = 3000
	ErrorSleepMs         = 5000
	EachLoopSleepMs      = 20000
)

func runHeartBeatCoroutine(this *AgwClient) {
	if this == nil {
		logWarn("AgwClient is null, exit heartbeat coroutine")
		return
	}

	for {
		for i := uint32(0); i < this.pool.size; i++ {
			conn, err := this.pool.get()

			if err != nil {
				logWarnf("get connection error:%s", err.Error())
				time.Sleep(time.Millisecond * ErrorSleepMs)
				continue
			}

			msg := NewAgwMessage()
			msg.SetReqId(generateId())
			msg.SetMessageType(MessageTypeHeartbeat)
			msg.SetMessageDirection(MessageDirectionRequest)
			msg.SetClientIp(StringIpToUint64(this.config.ClientIp))
			msg.SetClientVpcId(this.config.ClientVpcId)
			msg.SetServerName(HeartbeatServerName)
			msg.SetTimeoutMs(HeartbeatTimeoutMs)
			msg.SetClientProcessFlag(this.config.ClientProcessFlag)
			msg.SetConnectionId(conn.connId)
			msg.SetHandlerName(HeartbeatHandlerName)
			msg.SetOuterReqId("noReqIdForHB")
			msg.SetBody(HeartbeatMessageBody)

			err = conn.write(msg)
			if err != nil {
				time.Sleep(time.Millisecond * ErrorSleepMs)
				continue
			}
		}

		time.Sleep(time.Millisecond * EachLoopSleepMs)
	}

}
