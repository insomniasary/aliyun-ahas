package gateway

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func wait(conn *AgwConn, msg *AgwMessage) (*AgwMessage, error) {
	msgId := msg.getSyncId()
	channel := make(chan *AgwMessage, 1)

	_, loaded := conn.channels.LoadOrStore(msgId, channel)

	if loaded {
		close(channel)
		return nil, errors.New(ErrorMsgDupId)
	}

	ticker := time.NewTicker(req_timeout_sec * time.Second)
	defer ticker.Stop()

	select {
	case msg := <-channel:
		conn.channels.Delete(msgId)
		close(channel)

		if strings.Compare(msg.innerMsg, ErrorMsgConnClosed) == 0 {
			return nil, errors.New(ErrorMsgConnClosed)
		}
		return msg, nil
	case <-ticker.C:
		conn.channels.Delete(msgId)
		return nil, errors.New(fmt.Sprintf(ErrorMsgRequestTimeout))
	}
}

func notify(conn *AgwConn, msg *AgwMessage) {
	msgId := msg.getSyncId()

	channel, ok := conn.channels.Load(msgId)
	if !ok {
		logInfof("can not find channel by msgId:%s, connId: %d", msgId, conn.connId)
		return
	}

	conn.channels.Delete(msgId)
	if ch, ok := channel.(chan *AgwMessage); ok {
		ch <- msg
	}
}
