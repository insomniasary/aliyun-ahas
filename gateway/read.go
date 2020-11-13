package gateway

import (
	"bufio"
	"fmt"
)

func runReaderCoroutine(conn *AgwConn) {
	logInfof("AGW reader coroutine %d started", conn.connId)

	if conn == nil {
		logWarn("AGW conn is null, exit read coroutine")
		return
	}

	c := *(conn.conn)
	//*bufio.Reader
	bufReader := bufio.NewReaderSize(c, 128)

	for {
		msg := NewAgwMessage()
		if err := msg.Decode(bufReader); err != nil {
			if msg.ReqId() != 0 && msg.outerReqId != "" {
				logWarnf("AGW exit read coroutine, error:%s, reqId:%d, outerReqId:%s",
					err.Error(), msg.ReqId(), msg.OuterReqId())
			} else {
				logWarnf("AGW exit read coroutine, error:%s", err.Error())
			}
			conn.close()
			return
		}

		if msg.MessageType() == MessageTypeBiz && msg.MessageDirection() == MessageDirectionResponse {
			notify(conn, msg)
		} else if msg.MessageType() == MessageTypeBiz && msg.MessageDirection() == MessageDirectionRequest {
			handleRequest(msg, conn)
		} else if msg.MessageType() == MessageTypeHeartbeat && msg.MessageDirection() == MessageDirectionResponse {
			//todo : handle unexpected response case
		} else if msg.MessageType() == MessageTypeHeartbeat && msg.MessageDirection() == MessageDirectionRequest {
			msg.SetMessageDirection(MessageDirectionResponse)
			go conn.write(msg)
		} else {
			logWarnf("AGW unknown msg, type : %d, direction : %d, msg:%", msg.MessageType(), msg.MessageDirection(), msg)
		}

	}

}

func handleRequest(msg *AgwMessage, conn *AgwConn) {

	tsUtil := newTimestampUtil(msg.ReqId(), msg.OuterReqId())
	tsUtil.mark("gateway_call_client")

	handlerName := msg.HandlerName()
	handler, ok := handlers[handlerName]
	if !ok {
		logWarnf("AGW cannot get client handler by handlerName:%s, reqId:%d, outerReqId:%s",
			handlerName, msg.ReqId(), msg.OuterReqId())

		msg.SetInnerCode(8034)
		msg.SetInnerMsg("can not get client handler by handlerName")
		msg.SetMessageDirection(MessageDirectionResponse)

		go conn.write(msg)

		tsUtil.mark("no_handler_exception")
		logDebug(tsUtil.GetResult())

		return
	}

	tsUtil.mark("before_handle")
	response, err := handler.Handle(msg.Body())
	tsUtil.mark("after_handle")

	if err != nil {
		logWarnf("AGW executing client handler wrong, reqId:%d, outerReqId:%s", msg.ReqId(), msg.OuterReqId(), err.Error())

		msg.SetInnerCode(8035)
		msg.SetInnerMsg(fmt.Sprintf("executing client handler wrong : %s", err.Error()))
		msg.SetMessageDirection(MessageDirectionResponse)

		go conn.write(msg)

		tsUtil.mark("handle_exception")
		logDebug(tsUtil.GetResult())

		return
	}

	msg.SetMessageDirection(MessageDirectionResponse)
	msg.SetBody(response)

	go conn.write(msg)

	tsUtil.mark("write_ok")
	logDebug(tsUtil.GetResult())
}
