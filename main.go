package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	routers "github.com/kube-carbonara/cluster-agent/routers"
	"github.com/labstack/echo/v4"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
)

func init() {
}

var (
	addr  string
	id    string
	debug bool
)

func handleRouting(e *echo.Echo) {
	namespacesRouter := routers.NameSpacesRouter{}
	podsRouter := routers.PodsRouter{}
	deplymentRouter := routers.DeploymentsRouter{}
	serviceRouter := routers.SeviceRouter{}
	nodeRouter := routers.NodesRouter{}
	ingressRouter := routers.IngresRouter{}
	namespacesRouter.Handle(e)
	podsRouter.Handle(e)
	deplymentRouter.Handle(e)
	serviceRouter.Handle(e)
	nodeRouter.Handle(e)
	ingressRouter.Handle(e)
}

func main() {
	clusterGuid := os.Getenv("CLIENT_ID")
	flag.StringVar(&addr, "connect", fmt.Sprintf("ws://%s/connect", os.Getenv("SERVER_ADDRESS")), "Address to connect to")
	flag.StringVar(&id, "id", clusterGuid, "Client ID")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		"X-Tunnel-ID": []string{id},
	}
	time.AfterFunc(5*time.Second, func() {
		remotedialer.ClientConnect(context.Background(), addr, headers, nil, func(string, string) bool { return true },
			func(c context.Context) (err error) {
				if err != nil {
					fmt.Print("Failed to connect to the remote proxy restarting ...")
					os.Exit(3)
				}
				return err
			},
		)
	})

	e := echo.New()
	e.GET("/", func(context echo.Context) error {
		return context.String(http.StatusOK, "Hello, World!")
	})
	handleRouting(e)
	e.Logger.Fatal(e.Start(":1323"))
}
