package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/yejiabin9/pod/domain/model"
	"github.com/yejiabin9/pod/domain/repository"
	"github.com/yejiabin9/pod/proto/pod"
	v1 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

type IPodDataService interface {
	AddPod(pod *model.Pod) (int64, error)
	DeletePod(int64) error
	UpdatePod(*model.Pod) error
	FindPodByID(int64) (*model.Pod, error)
	FindAllPod() ([]model.Pod, error)
	CreateToK8s(*pod.PodInfo) error
	DeleteFromK8s(*model.Pod) error
	UpdateToK8s(*pod.PodInfo) error
}

func NewPodDataService(podRepository repository.IPodRepository, clientset *kubernetes.Clientset) IPodDataService {
	return &PodDataService{
		PodRepository: podRepository,
		K8sClientSet:  clientset,
		deployment:    &v1.Deployment{},
	}
}

type PodDataService struct {
	PodRepository repository.IPodRepository
	K8sClientSet  *kubernetes.Clientset
	deployment    *v1.Deployment
}

func (u *PodDataService) AddPod(pod *model.Pod) (int64, error) {
	return u.PodRepository.CreatePod(pod)
}

func (u *PodDataService) DeletePod(i int64) error {
	return u.PodRepository.DeletePodByID(i)
}

func (u *PodDataService) UpdatePod(m *model.Pod) error {
	return u.PodRepository.UpdatePod(m)
}

func (u *PodDataService) FindPodByID(i int64) (*model.Pod, error) {
	return u.PodRepository.FindPodByID(i)
}

func (u *PodDataService) FindAllPod() ([]model.Pod, error) {
	return u.PodRepository.FindAll()
}

func (u *PodDataService) CreateToK8s(podInfo *pod.PodInfo) error {
	u.SetDeployment(podInfo)
	if _, err := u.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Get(context.TODO(), podInfo.PodName, v12.GetOptions{}); err != nil {
		if _, err = u.K8sClientSet.AppsV1().Deployments(podInfo.PodNamespace).Create(context.TODO(), u.deployment, v12.CreateOptions{}); err != nil {
			fmt.Println(err.Error())
			return err
		}
		//common.Info("创建成功")
		fmt.Println("创建成功")
		return nil
	} else {
		//可以写自己的业务逻辑
		fmt.Println("Pod" + podInfo.PodName + "已经存在")
		//common.Error("Pod " + podInfo.PodName + "已经存在")
		return errors.New("Pod " + podInfo.PodName + " 已经存在")
	}
}

func (u PodDataService) DeleteFromK8s(pod *model.Pod) (err error) {
	if err = u.K8sClientSet.AppsV1().Deployments(pod.PodNamespace).Delete(context.TODO(), pod.PodName, v12.DeleteOptions{}); err != nil {
		fmt.Println(err.Error())
		//写自己的业务逻辑
		return err
	} else {
		if err := u.DeletePod(pod.ID); err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Println("删除Pod ID：" + strconv.FormatInt(pod.ID, 10) + " 成功！")
		//common.Info("删除Pod ID：" + strconv.FormatInt(pod.ID, 10) + " 成功！")
	}
	return
}

func (u *PodDataService) UpdateToK8s(info *pod.PodInfo) (err error) {
	u.SetDeployment(info)
	if _, err = u.K8sClientSet.AppsV1().Deployments(info.PodNamespace).Get(context.TODO(), info.PodName, v12.GetOptions{}); err != nil {
		fmt.Println(err.Error())
		return errors.New("Pod " + info.PodName + " 不存在请先创建")
	} else {
		//如果存在
		if _, err = u.K8sClientSet.AppsV1().Deployments(info.PodNamespace).Update(context.TODO(), u.deployment, v12.UpdateOptions{}); err != nil {
			fmt.Println(err.Error())
			return err
		}
		fmt.Println(info.PodName + " 更新成功")
		return nil
	}

}

func (u *PodDataService) SetDeployment(podInfo *pod.PodInfo) {
	deployment := &v1.Deployment{}
	deployment.TypeMeta = v12.TypeMeta{
		Kind:       "deployment",
		APIVersion: "v1",
	}
	deployment.ObjectMeta = v12.ObjectMeta{
		Name:      podInfo.PodName,
		Namespace: podInfo.PodNamespace,
		Labels: map[string]string{
			"app-name": podInfo.PodName,
			"author":   "Caplost",
		},
	}
	deployment.Name = podInfo.PodName
	deployment.Spec = v1.DeploymentSpec{
		//副本个数
		Replicas: &podInfo.PodReplicas,
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app-name": podInfo.PodName,
			},
			MatchExpressions: nil,
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Labels: map[string]string{
					"app-name": podInfo.PodName,
				},
			},
			Spec: v13.PodSpec{
				Containers: []v13.Container{
					{
						Name:            podInfo.PodName,
						Image:           podInfo.PodImage,
						Ports:           u.getContainerPort(podInfo),
						Env:             u.getEnv(podInfo),
						Resources:       u.getResources(podInfo),
						ImagePullPolicy: u.getImagePullPolicy(podInfo),
					},
				},
			},
		},
		Strategy:                v1.DeploymentStrategy{},
		MinReadySeconds:         0,
		RevisionHistoryLimit:    nil,
		Paused:                  false,
		ProgressDeadlineSeconds: nil,
	}
	u.deployment = deployment
}

func (u *PodDataService) getContainerPort(podInfo *pod.PodInfo) (containerPort []v13.ContainerPort) {
	for _, v := range podInfo.PodPort {
		containerPort = append(containerPort, v13.ContainerPort{
			Name:          "port-" + strconv.FormatInt(int64(v.ContainerPort), 10),
			ContainerPort: v.ContainerPort,
			Protocol:      u.getProtocol(v.Protocol),
		})
	}
	return
}

func (u *PodDataService) getImagePullPolicy(podInfo *pod.PodInfo) v13.PullPolicy {
	switch podInfo.PodPollPolicy {
	case "Always":
		return "Always"
	case "Never":
		return "Never"
	case "IfNotPresent":
		return "IfNotPresent"
	default:
		return "Always"
	}
}

func (u *PodDataService) getProtocol(protocol string) v13.Protocol {
	switch protocol {
	case "TCP":
		return "TCP"
	case "UDP":
		return "UDP"
	case "SCTP":
		return "SCTP"
	default:
		return "TCP"
	}
}

func (u *PodDataService) getEnv(podInfo *pod.PodInfo) (envVar []v13.EnvVar) {
	for _, v := range podInfo.PodEnv {
		envVar = append(envVar, v13.EnvVar{
			Name:      v.EnvKey,
			Value:     v.EnvValue,
			ValueFrom: nil,
		})
	}
	return
}

func (u *PodDataService) getResources(podInfo *pod.PodInfo) (source v13.ResourceRequirements) {
	//最大能够使用多少资源
	source.Limits = v13.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax), 'f', 6, 64)),
	}
	//满足最少使用的资源量
	//@TODO 自己实现动态设置
	source.Requests = v13.ResourceList{
		"cpu":    resource.MustParse(strconv.FormatFloat(float64(podInfo.PodCpuMax), 'f', 6, 64)),
		"memory": resource.MustParse(strconv.FormatFloat(float64(podInfo.PodMemoryMax), 'f', 6, 64)),
	}
	return
}
