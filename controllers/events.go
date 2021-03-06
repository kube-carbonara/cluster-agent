package controllers

import (
	ctx "context"
	"log"
	"net/http"
	"time"

	"github.com/kube-carbonara/cluster-agent/models"
	services "github.com/kube-carbonara/cluster-agent/services"
	utils "github.com/kube-carbonara/cluster-agent/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	CoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type EventsController struct {
}

func (c EventsController) runWatcherEventLoop() error {
	config := utils.NewConfig()
	var client utils.Client = *utils.NewClient()
	watch, err := client.Clientset.CoreV1().Events(CoreV1.NamespaceAll).Watch(ctx.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
		return err
	}

	channel := watch.ResultChan()

	done := make(chan struct{})

	session := utils.Session{
		Host:    config.RemoteProxy,
		Channel: "monitoring",
	}
	session.NewSession()
	defer session.Conn.Close()
	defer close(done)
	for {
		select {
		case event, ok := <-channel:
			if !ok {
				log.Fatal("unexpected type")
				return nil
			}

			obj, ok := event.Object.(*CoreV1.Event)
			if !ok {
				log.Fatal("unexpected type")
				return nil
			}
			err := services.MonitoringService{
				NameSpace: obj.Namespace,
				EventName: string(event.Type),
				Resource:  utils.EVENTS,
				PayLoad:   obj,
			}.PushEvent(&session)

			if err != nil {
				logrus.Error(err)
				session.Conn.Close()
				session = *session.NewSession()
				services.MonitoringService{
					EventName: string(event.Type),
					Resource:  utils.EVENTS,
					PayLoad:   obj,
				}.PushEvent(&session)
			}

		case <-time.After(30 * time.Minute):
			logrus.Info("Timeout, restarting event watcher")
			return nil

		}
	}

}

func (c EventsController) Watch() {
	for {
		if err := c.runWatcherEventLoop(); err != nil {
			logrus.Error(err)
		}

	}

}

func (c EventsController) GetOne(context echo.Context, name string, nameSpace string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.CoreV1().Events(nameSpace).Get(ctx.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.EVENTS,
	})
}

func (c EventsController) Get(context echo.Context, nameSpace string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.CoreV1().Events(nameSpace).List(ctx.TODO(), metav1.ListOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.EVENTS,
	})
}
