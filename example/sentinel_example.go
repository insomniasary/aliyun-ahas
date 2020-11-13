package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/util"
	ahas "github.com/aliyun/aliyun-ahas-go-sdk"
)

func main() {
	// Note: set the config path via SENTINEL_CONFIG_FILE_PATH system env.
	err := ahas.InitAhasDefault()
	if err != nil {
		log.Fatalf("Failed to init AHAS: %+v", err)
	}

	playSentinel()
}

type stateChangeTestListener struct {
}

func (s *stateChangeTestListener) OnTransformToClosed(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("cb.strategy: %+v, From %s to Closed, time: %d\n", rule.Strategy.String(), prev.String(), util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnTransformToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	fmt.Printf("cb.strategy: %+v, From %s to Open, snapshot: %.2f, time: %d\n", rule.Strategy.String(), prev.String(), snapshot, util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnTransformToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("cb.strategy: %+v, From %s to Half-Open, time: %d\n", rule.Strategy.String(), prev.String(), util.CurrentTimeMillis())
}

func goWithEntry(res string, t base.TrafficType, c base.ResourceType, maxSleepMs uint32) {
	go func() {
		for {
			e, b := sentinel.Entry(res, sentinel.WithTrafficType(t), sentinel.WithResourceType(c))
			if b != nil {
				time.Sleep(time.Duration(rand.Uint32()%maxSleepMs+1) * time.Millisecond)
			} else {
				// Passed, wrap the logic here.
				time.Sleep(time.Duration(rand.Uint32()%maxSleepMs+1) * time.Millisecond)
				e.Exit()
			}
		}
	}()
}

func playSentinel() {
	rand.Seed(time.Now().UnixNano())
	ch := make(chan struct{})
	circuitbreaker.RegisterStateChangeListeners(&stateChangeTestListener{})

	goWithEntry("GET:/foo/:id", base.Inbound, base.ResTypeWeb, 7)
	goWithEntry("/grpc.testing.TestService/FooCall", base.Inbound, base.ResTypeRPC, 20)

	for i := 0; i < 8; i++ {
		goWithEntry("SELECT * FROM user WHERE id = ?", base.Outbound, base.ResTypeDBSQL, 15)
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
		go func() {
			for {
				e, b := sentinel.Entry("order-service", sentinel.WithTrafficType(base.Inbound), sentinel.WithArgs(rand.Uint32()%10))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}
