package main

import (
	"flag"
	"fmt"
	"github.com/asim/go-micro/plugins/registry/consul/v3"
	"github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/registry"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	"github.com/yejiabin9/common"
	"github.com/yejiabin9/pod/domain/repository"
	service2 "github.com/yejiabin9/pod/domain/service"
	"github.com/yejiabin9/pod/handler"
	"github.com/yejiabin9/pod/proto/pod"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"strconv"
)

var (
	consulHost       = "39.104.82.215"
	consulPort int64 = 8500

	tracerHost       = ""
	tracerPort int64 = 6831

	hystrixPort int64 = 9092

	prometheusPort int64 = 9192
)

func main() {
	consul := consul.NewRegistry(func(options *registry.Options) {
		options.Addrs = []string{
			consulHost + ":" + strconv.FormatInt(consulPort, 10),
		}
	})

	consulConfig, err := common.GetConsulConfig(consulHost, consulPort, "micro/config")
	if err != nil {
		logrus.Error(err.Error())
	}

	//connect mysql
	mysqlInfo := common.GetMysqlFromConsul(consulConfig, "mysql")
	//init sql
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local&timeout=%s", mysqlInfo.User, mysqlInfo.Pwd, mysqlInfo.Host, 3306, mysqlInfo.Database, "10s")
	//open, err := gorm.Open(
	//	mysql.Open(dsn),
	//	&gorm.Config{})
	//if err != nil {
	//	logrus.Error("connect mysql error")
	//	return
	//}
	//db, _ := open.DB()
	//defer db.Close()
	//db, err := gorm.Open(mysql.New(mysql.Config{
	//	DSN: dsn,
	//}), &gorm.Config{
	//	SkipDefaultTransaction: false, //跳过默认事务
	//	NamingStrategy: schema.NamingStrategy{
	//		SingularTable: true, // 复数形式 User的表名应该是users
	//		TablePrefix:   "t_", //表名前缀 User的表名应该是t_users
	//	},
	//	DisableForeignKeyConstraintWhenMigrating: true, //设置成为逻辑外键(在物理数据库上没有外键，仅体现在代码上)
	//})
	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//————————————————
	//版权声明：本文为CSDN博主「童话ing」的原创文章，遵循CC 4.0 BY-SA版权协议，转载请附上原文出处链接及本声明。
	//原文链接：https://blog.csdn.net/dl962454/article/details/124109828

	//t, io, err := common..newT

	//download kubectl
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"),
			"kubeconfgi file 在当前系统地址中")
	} else {
		kubeConfig = flag.String("kubeconfig", "", "kubeconfgi file 在当前系统地址中")
	}

	flag.Parse()

	//create config 实例
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		logrus.Error(err.Error())
	}
	//config, err := rest.InClusterConfig()

	//create k8s client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Error(err.Error())
	}

	//create service
	service := micro.NewService(
		micro.Name("go.micro.service.pod"),
		micro.Version("latest"),
		micro.Registry(consul),
	)

	//init service

	service.Init()

	//create table
	err = repository.NewPodRepository(db).InitTable()
	if err != nil {
		logrus.Error(err.Error())
	}

	podDataService := service2.NewPodDataService(repository.NewPodRepository(db), clientset)
	pod.RegisterPodHandler(service.Server(), &handler.PodHandler{PodDataService: podDataService})

	err = service.Run()
	if err != nil {
		logrus.Error(err.Error())
	}
}
