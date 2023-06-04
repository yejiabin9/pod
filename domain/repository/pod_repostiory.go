package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/yejiabin9/pod/domain/model"
)

type IPodRepository interface {
	//初始化表
	InitTable() error
	FindPodByID(int64) (*model.Pod, error)

	CreatePod(*model.Pod) (int64, error)

	DeletePodByID(int64) error

	UpdatePod(*model.Pod) error

	FindAll() ([]model.Pod, error)
}

func NewPodRepository(db *gorm.DB) IPodRepository {

	return &PodRepository{mysqlDb: db}

}

type PodRepository struct {
	mysqlDb *gorm.DB
}

func (u *PodRepository) InitTable() error {
	return u.mysqlDb.CreateTable(&model.Pod{}, &model.PodEnv{}, &model.PodPort{}).Error
}

func (u *PodRepository) FindPodByID(podId int64) (pod *model.Pod, err error) {
	pod = &model.Pod{}
	return pod, u.mysqlDb.Preload("PodEnv").Preload("PodPort").First(pod, podId).Error
}

func (u *PodRepository) CreatePod(pod *model.Pod) (int64, error) {
	return pod.ID, u.mysqlDb.Create(pod).Error

}

func (u *PodRepository) DeletePodByID(podID int64) error {
	tx := u.mysqlDb.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}

	}()

	if tx.Error != nil {
		return tx.Error
	}

	if err := u.mysqlDb.Where("id = ?", podID).Delete(&model.Pod{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	//delete podEnv
	if err := u.mysqlDb.Where("id = ?", podID).Delete(&model.PodEnv{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	//delete port
	if err := u.mysqlDb.Where("id = ?", podID).Delete(&model.PodPort{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error

}

func (u *PodRepository) UpdatePod(pod *model.Pod) error {
	return u.mysqlDb.Model(pod).Update(pod).Error
}

func (u *PodRepository) FindAll() (podAll []model.Pod, err error) {
	return podAll, u.mysqlDb.Find(&podAll).Error

}
