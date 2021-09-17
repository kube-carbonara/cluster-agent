package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/kube-carbonara/cluster-agent/controllers"
	routers "github.com/kube-carbonara/cluster-agent/routers"
	"github.com/kube-carbonara/cluster-agent/utils"
	"github.com/labstack/echo/v4"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

func init() {
}

var (
	addr   string
	id     string
	debug  bool
	appKey string
)

func handleRouting(e *echo.Echo) {
	namespacesRouter := routers.NameSpacesRouter{}
	podsRouter := routers.PodsRouter{}
	deplymentRouter := routers.DeploymentsRouter{}
	serviceRouter := routers.SeviceRouter{}
	nodeRouter := routers.NodesRouter{}
	ingressRouter := routers.IngresRouter{}
	metricsRouter := routers.MetricsRouter{}
	namespacesRouter.Handle(e)
	podsRouter.Handle(e)
	deplymentRouter.Handle(e)
	serviceRouter.Handle(e)
	nodeRouter.Handle(e)
	ingressRouter.Handle(e)
	metricsRouter.Handle(e)
}

func main() {
	config := utils.NewConfig()
	flag.StringVar(&addr, "connect", fmt.Sprintf("ws://%s/connect", config.RemoteProxy), "Address to connect to")
	flag.StringVar(&id, "id", config.ClientId, "Client ID")
	flag.StringVar(&appKey, "appKey", config.AppKey, "App Key")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		"X-Tunnel-ID":     []string{id},
		"x-agent":         []string{id},
		"x-agent-app-key": []string{appKey},
	}
	time.AfterFunc(5*time.Second, func() {
		remotedialer.ClientConnect(context.Background(), addr, headers, nil, func(string, string) bool { return true }, nil)

	})

	session := utils.Session{
		Host:    config.RemoteProxy,
		Channel: "monitoring",
	}
	session.NewSession()
	defer session.Conn.Close()
	controllers.ServicesController{}.Watch(&session)
	controllers.PodsController{}.Watch(&session)
	controllers.DeploymentsController{}.Watch(&session)
	controllers.NameSpacesController{}.Watch(&session)
	controllers.NodesController{}.Watch(&session)
	controllers.IngressController{}.Watch(&session)

	e := echo.New()
	e.GET("/", func(context echo.Context) error {
		return context.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/health", func(context echo.Context) error {
		return context.String(http.StatusOK, "app is running")
	})
	handleRouting(e)

	e.Logger.Fatal(e.Start(":1323"))
}
