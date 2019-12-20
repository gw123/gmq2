package core

import (
	"github.com/gw123/gmq/core/interfaces"

)

type MiddlewareManager struct {
	middlewares []interfaces.Middleware
	app         interfaces.App
}

func NewMiddlewareManager(app interfaces.App) *MiddlewareManager {
	this := new(MiddlewareManager)
	this.app = app
	return this
}

func (this *MiddlewareManager) Handel(event interfaces.Msg) bool {
	for _, middleware := range this.middlewares {
		flag := middleware.Handel(event)
		if !flag {
			return false
		}
	}
	return true
}

func (this *MiddlewareManager) RegisterMiddleware(middleware interfaces.Middleware) {
	this.middlewares = append(this.middlewares, middleware)
}
