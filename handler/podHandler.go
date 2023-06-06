package handler

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/yejiabin9/pod/domain/model"
	"github.com/yejiabin9/pod/domain/service"
	"github.com/yejiabin9/pod/proto/pod"
	"github.com/yejiabin9/pod/utils"
	"strconv"
)

type PodHandler struct {
	PodDataService service.IPodDataService
}

func (e *PodHandler) AddPod(ctx context.Context, info *pod.PodInfo, rsp *pod.Response) error {
	logrus.Info("add pod")
	podModel := &model.Pod{}
	err := utils.SwapTo(info, podModel)

	if err != nil {
		logrus.Error(err.Error())
		rsp.Msg = err.Error()
		return err
	}

	if err := e.PodDataService.CreateToK8s(info); err != nil {
		logrus.Error(err.Error())
		rsp.Msg = err.Error()
		return err
	} else {
		podID, err := e.PodDataService.AddPod(podModel)
		if err != nil {
			logrus.Error(err.Error())
			rsp.Msg = err.Error()
			return err
		}
		logrus.Info("create pod success, podID is ", podID)
		rsp.Msg = "Pod create success. podID is " + strconv.FormatInt(podID, 10)
	}
	return nil
}

func (e *PodHandler) DeletePod(ctx context.Context, podId *pod.PodID, rsp *pod.Response) error {
	podModle, err := e.PodDataService.FindPodByID(int64(podId.PodId))
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	if err := e.PodDataService.DeleteFromK8s(podModle); err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}

func (e *PodHandler) FindPodByID(ctx context.Context, id *pod.PodID, rsp *pod.PodInfo) error {
	podModel, err := e.PodDataService.FindPodByID(int64(id.PodId))
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	err = utils.SwapTo(podModel, rsp)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	return nil
}

func (e *PodHandler) UpdatePod(ctx context.Context, info *pod.PodInfo, response *pod.Response) error {
	err := e.PodDataService.UpdateToK8s(info)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	podModel, err := e.PodDataService.FindPodByID(info.Id)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	err = utils.SwapTo(info, podModel)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	e.PodDataService.UpdatePod(podModel)
	return nil
}

func (e *PodHandler) FindAllPod(ctx context.Context, req *pod.FindAll, rsp *pod.AllPod) error {
	allPod, err := e.PodDataService.FindAllPod()
	if err != nil {
		logrus.Error(err.Error())
		return err
	}
	for _, v := range allPod {
		podInfo := &pod.PodInfo{}
		err := utils.SwapTo(v, podInfo)
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		rsp.PodInfo = append(rsp.PodInfo, podInfo)
	}
	return nil
}
