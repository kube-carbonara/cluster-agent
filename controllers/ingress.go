package controllers

import (
	ctx "context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/kube-carbonara/cluster-agent/models"
	services "github.com/kube-carbonara/cluster-agent/services"
	utils "github.com/kube-carbonara/cluster-agent/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	CoreV1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressController struct {
}

func (c IngressController) WatchTest(session *utils.Session) {
	go func() {

		for {
			services.MonitoringService{}.PushEvent(session)
			time.Sleep(30 * time.Second)
		}

	}()

}

func (c IngressController) runWatcherEventLoop() error {
	config := utils.NewConfig()
	var client utils.Client = *utils.NewClient()
	watch, err := client.Networkingv1client.Ingresses(CoreV1.NamespaceAll).Watch(ctx.TODO(), metav1.ListOptions{})
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
			obj, ok := event.Object.(*networkingv1.Ingress)
			if !ok {
				log.Fatal("unexpected type")
				return nil
			}

			err := services.MonitoringService{
				NameSpace: obj.Namespace,
				EventName: string(event.Type),
				Resource:  utils.RESOUCETYPE_INGRESS,
				PayLoad:   obj,
			}.PushEvent(&session)

			if err != nil {
				logrus.Error(err)
				session.Conn.Close()
				session = *session.NewSession()
				services.MonitoringService{
					EventName: string(event.Type),
					Resource:  utils.RESOUCETYPE_INGRESS,
					PayLoad:   obj,
				}.PushEvent(&session)
			}

		case <-time.After(30 * time.Minute):
			logrus.Info("Timeout, restarting event watcher")
			return nil

		}
	}

}

func (c IngressController) Watch() {
	for {
		if err := c.runWatcherEventLoop(); err != nil {
			logrus.Error(err)
		}

	}
}

func (c IngressController) GetOne(context echo.Context, nameSpaceName string, name string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Networkingv1client.Ingresses(nameSpaceName).Get(ctx.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_INGRESS,
	})
}

func (c IngressController) Get(context echo.Context, nameSpaceName string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Networkingv1client.Ingresses(nameSpaceName).List(ctx.TODO(), metav1.ListOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_INGRESS,
	})
}

func (c IngressController) Create(context echo.Context, nameSpaceName string, ingressConfig map[string]interface{}) error {
	ingress := &networkingv1.Ingress{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(ingressConfig), ingress)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}

	var client utils.Client = *utils.NewClient()
	ingress, err := client.Networkingv1client.Ingresses(nameSpaceName).Create(ctx.TODO(), ingress, metav1.CreateOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}
	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(ingress),
		ResourceType: utils.RESOUCETYPE_INGRESS,
	})
}

func (c IngressController) Update(context echo.Context, nameSpaceName string, ingressConfig map[string]interface{}) error {
	ingress := &networkingv1.Ingress{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(ingressConfig), ingress)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}
	var client utils.Client = *utils.NewClient()
	ingress, err := client.Networkingv1client.Ingresses(nameSpaceName).Update(ctx.TODO(), ingress, metav1.UpdateOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}
	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(ingress),
		ResourceType: utils.RESOUCETYPE_INGRESS,
	})
}

func (c IngressController) Delete(context echo.Context, nameSpaceName string, name string) error {
	var client utils.Client = *utils.NewClient()
	err := client.Networkingv1client.Ingresses(nameSpaceName).Delete(ctx.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}
	return context.JSON(http.StatusNoContent, nil)
}
