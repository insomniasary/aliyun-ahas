package service

import (
	"context"
	"sync"
)

type LifeCycle interface {
	Start() error
	Stop() error
}

type LifeCycle0 interface {
	DoStart() error
	DoStop() error
}

type Controller struct {
	mutex0 sync.Mutex
	Ctx    context.Context
	Cancel context.CancelFunc
	LifeCycle
	LifeCycle0
	IsStarted bool
}

//NewController
func NewController(cycle LifeCycle0) *Controller {
	controller := &Controller{
		mutex0: sync.Mutex{},
	}
	controller.LifeCycle = controller
	controller.LifeCycle0 = cycle
	return controller
}

//Start
func (controller *Controller) Start() error {
	controller.mutex0.Lock()
	defer controller.mutex0.Unlock()
	if controller.Ctx == nil || controller.Ctx.Err() != nil {
		ctx, cancel := context.WithCancel(context.Background())
		controller.Ctx = ctx
		controller.Cancel = cancel
		err := controller.DoStart()
		controller.IsStarted = true
		return err
	}
	return nil
}

//Stop
func (controller *Controller) Stop() error {
	controller.mutex0.Lock()
	defer controller.mutex0.Unlock()
	if controller.Ctx != nil && controller.Ctx.Err() == nil {
		controller.Cancel()
		err := controller.DoStop()
		controller.IsStarted = false
		return err
	}
	return nil
}

func (controller *Controller) IsStopped() bool {
	return controller.Ctx == nil || controller.Ctx.Err() != nil
}
