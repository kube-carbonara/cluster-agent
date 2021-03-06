package controllers

import (
	ctx "context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/kube-carbonara/cluster-agent/models"
	services "github.com/kube-carbonara/cluster-agent/services"
	utils "github.com/kube-carbonara/cluster-agent/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	CoreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type DeploymentsController struct{}

func (c DeploymentsController) WatchTest(session *utils.Session) {
	go func() {

		for {
			err := services.MonitoringService{}.PushEvent(session)
			if err != nil {
				logrus.Error("Error sending deployment events: ", err.Error())
			}
		}

	}()

}

func (c DeploymentsController) runWatcherEventLoop() error {
	config := utils.NewConfig()
	var client utils.Client = *utils.NewClient()
	watch, err := client.Clientset.AppsV1().Deployments(CoreV1.NamespaceAll).Watch(ctx.TODO(), metav1.ListOptions{})
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
			obj, ok := event.Object.(*v1.Deployment)
			if !ok {
				log.Fatal("unexpected type")
				return nil
			}

			err := services.MonitoringService{
				NameSpace: obj.Namespace,
				EventName: string(event.Type),
				Resource:  utils.RESOUCETYPE_DEPLOYMENTS,
				PayLoad:   obj,
			}.PushEvent(&session)

			if err != nil {
				logrus.Error(err)
				session.Conn.Close()
				session = *session.NewSession()
				services.MonitoringService{
					EventName: string(event.Type),
					Resource:  utils.RESOUCETYPE_DEPLOYMENTS,
					PayLoad:   obj,
				}.PushEvent(&session)
			}

		case <-time.After(30 * time.Minute):
			logrus.Info("Timeout, restarting event watcher")
			return nil

		}
	}

}

func (c DeploymentsController) Watch() error {
	for {
		if err := c.runWatcherEventLoop(); err != nil {
			logrus.Error(err)
		}

	}
}
func (c DeploymentsController) GetOne(context echo.Context, nameSpaceName string, name string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).Get(ctx.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) Get(context echo.Context, nameSpaceName string) error {
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).List(ctx.TODO(), metav1.ListOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) GetBySelector(context echo.Context, nameSpaceName string, selector string) error {
	labelSelector := c.parseSelector(selector)
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).List(ctx.TODO(), metav1.ListOptions{
		LabelSelector: labels.SelectorFromSet(labelSelector).String(),
	})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) Create(context echo.Context, nameSpaceName string, deploymentConfig map[string]interface{}) error {
	deployment := &v1.Deployment{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(deploymentConfig), deployment)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}
	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).Create(ctx.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) Update(context echo.Context, nameSpaceName string, deploymentConfig map[string]interface{}) error {
	deployment := &v1.Deployment{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(deploymentConfig), deployment)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}

	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).Update(ctx.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) Delete(context echo.Context, nameSpaceName string, name string) error {
	var client utils.Client = *utils.NewClient()
	err := client.Clientset.AppsV1().Deployments(nameSpaceName).Delete(ctx.TODO(), name, metav1.DeleteOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusNoContent, models.Response{
		Data:         nil,
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) Restart(context echo.Context, nameSpaceName string, deploymentConfig map[string]interface{}) error {
	deployment := &v1.Deployment{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(deploymentConfig), deployment)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}
	deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)

	if deployment.Spec.Template.ObjectMeta.Annotations == nil {
		deployment.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

	var client utils.Client = *utils.NewClient()
	result, err := client.Clientset.AppsV1().Deployments(nameSpaceName).Update(ctx.TODO(), deployment, metav1.UpdateOptions{})
	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}

	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) ReScale(context echo.Context, nameSpaceName string, scale int32, deploymentConfig map[string]interface{}) error {
	deployment := &v1.Deployment{}
	UnmarshalErr := json.Unmarshal(utils.MapToJson(deploymentConfig), deployment)
	if UnmarshalErr != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: UnmarshalErr.Error(),
		})
	}
	var client utils.Client = *utils.NewClient()
	s, err := client.Clientset.AppsV1().
		Deployments(nameSpaceName).
		GetScale(ctx.TODO(), deployment.ObjectMeta.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	sc := *s
	sc.Spec.Replicas = scale

	result, err := client.Clientset.AppsV1().
		Deployments(nameSpaceName).
		UpdateScale(ctx.TODO(),
			deployment.ObjectMeta.Name, &sc, metav1.UpdateOptions{})

	if err != nil {
		return context.JSON(http.StatusBadRequest, models.Response{
			Message: err.Error(),
		})
	}
	return context.JSON(http.StatusOK, models.Response{
		Data:         utils.StructToMap(result),
		ResourceType: utils.RESOUCETYPE_DEPLOYMENTS,
	})
}

func (c DeploymentsController) parseSelector(selector string) labels.Set {
	specSelector := map[string]string{}
	selectors := strings.Split(selector, ";")
	for _, v := range selectors {
		if s := strings.Split(v, "="); len(s) > 1 {
			key := s[0]
			value := s[1]
			specSelector[key] = value
		}

	}
	set := labels.Set(specSelector)
	return set

}
