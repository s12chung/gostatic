package router

import (
	logTest "github.com/sirupsen/logrus/hooks/test"

	"github.com/sirupsen/logrus"
)

type GenerateRouterSetup struct{}

func NewGenerateRouterSetup() *GenerateRouterSetup {
	return &GenerateRouterSetup{}
}

func (setup *GenerateRouterSetup) DefaultRouter() (Router, logrus.FieldLogger, *logTest.Hook) {
	return defaultGenerateRouter()
}

func (setup *GenerateRouterSetup) RunServer(router Router, callback func()) {
	callback()
}

func (setup *GenerateRouterSetup) Requester(router Router) Requester {
	return router.Requester()
}

func defaultGenerateRouter() (*GenerateRouter, logrus.FieldLogger, *logTest.Hook) {
	log, hook := logTest.NewNullLogger()
	return NewGenerateRouter(log), log, hook
}
