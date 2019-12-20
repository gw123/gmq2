package core

import (
	"github.com/gw123/gmq/core/interfaces"
	"sync"
	"strings"
	"time"
)

type Dispatch struct {
	EventQueueBinds map[string][]interfaces.Module
	EventQueues     interfaces.EventQueue
	app             interfaces.App
	mutex           sync.Mutex
	appEventNames   []string
}

func NewDispath(app interfaces.App) *Dispatch {
	this := new(Dispatch)
	this.app = app
	this.EventQueueBinds = make(map[string][]interfaces.Module)
	this.EventQueues = NewEventQueue(app)
	return this
}

func (this *Dispatch) SetEventNames(eventNames string) {
	this.appEventNames = strings.Split(eventNames, ",")
}

func (this *Dispatch) Start() {
	for ; ; {
		event, err := this.EventQueues.Pop()
		//if event != nil {
		//	this.app.Debug("Dispath", "Start pop:"+event.GetMsgId()+" : "+event.GetEventName())
		//}
		if err != nil {
			if err.Error() == "队列为空" {

			} else {
				this.app.Warn("app", "出队异常:"+err.Error())
			}
			time.Sleep(time.Millisecond)
			continue
		}

		if event == nil {
			this.app.Warn("app", "出队异常:event为nil")
			continue
		}

		eventName := event.GetEventName()
		modules := this.EventQueueBinds[eventName]
		//this.app.Debug("Dispatch",fmt.Sprintf("Bingdings modules len %d", len(modules)))
		this.app.Handel(event)
		for _, module := range modules {
			//this.app.Debug("Dispatch",fmt.Sprintf("Bingding moduleName %s", module.GetModuleName()))
			err := module.Push(event)
			if err != nil {
				this.app.Warn(module.GetModuleName(), "模块队列异常Push失败"+err.Error())
			}
		}
	}
}

func (this *Dispatch) PushToModule(event interfaces.Msg) {
	eventName := event.GetEventName()
	modules := this.EventQueueBinds[eventName]
	//this.app.Debug("Dispatch",fmt.Sprintf("Bingdings modules len %d", len(modules)))
	for _, module := range modules {
		//this.app.Debug("Dispatch",fmt.Sprintf("Bingding moduleName %s", module.GetModuleName()))
		err := module.Push(event)
		if err != nil {
			this.app.Warn(module.GetModuleName(), "模块队列异常Push失败"+err.Error())
		}
	}
}

func (this *Dispatch) handel(event interfaces.Msg) {
	app, ok := this.app.(*App)
	if ok {
		app.Handel(event)
	}
}

func (this *Dispatch) Sub(eventName string, module interfaces.Module) {
	if eventName == "" || eventName == " " {
		this.app.Warn("Dispatch", "Sub eventName 为空,"+" moduleName: "+module.GetModuleName())
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	modules := this.EventQueueBinds[eventName]
	for _, m := range modules {
		if m.GetModuleName() == module.GetModuleName() {
			this.app.Warn("sub", m.GetModuleName()+"已经订阅"+eventName)
			return
		}
	}
	this.EventQueueBinds[eventName] = append(this.EventQueueBinds[eventName], module)
}

func (this *Dispatch) UnSub(eventName string, module interfaces.Module) {
	if eventName == "" {
		this.app.Warn("Dispatch", "UnSub eventName 为空")
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	modules := this.EventQueueBinds[eventName]
	for index, m := range modules {
		if m.GetModuleName() == module.GetModuleName() {
			modules[index] = nil
			this.app.Info("sub", m.GetModuleName()+"取消订阅"+eventName)
			return
		}
	}
}

func (this *Dispatch) Pub(event interfaces.Msg) {
	if event.GetEventName() == "" {
		this.app.Warn("Dispatch", "Pub eventName 为空"+"moduleName:"+event.GetEventName()+"srouceModule"+event.GetSourceModule())
		return
	}
	//this.app.Debug("Dispath", event.GetMsgId()+" : "+event.GetEventName())
	this.EventQueues.Push(event)
	//this.PushToModule(event)
}
