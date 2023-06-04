package model

type Pod struct {
	ID           int64  `gorm:"primary_key;not_null;auto_increment" json:"id"`
	PodName      string `gorm:"unique_index;not_null" json:"pod_name"`
	PodNamespace string `json:"pod_namespace"`

	//pod团队
	PodTeamID int64 `json:"pod_team_id"`

	PodCpuMin float32 `json:"pod_cpu_min"`
	PodCpuMax float32 `json:"pod_cpu_max"`

	PodReplicas int32 `json:"pod_replicas"`

	PodMemoryMin float32 `json:"pod_memory_min"`
	PodMemoryMax float32 `json:"pod_memory_max"`

	PodPort []PodPort `gorm:"ForeignKey:PodID" json:"pod_port"`
	PodEnv  []PodEnv  `gorm:"ForeignKey:PodID" json:"pod_env"`

	//镜像拉取策略
	PodPullPolicy string `json:"pod_pull_policy"`

	//pod重启策略
	PodRestart string `json:"pod_restart"`

	//发布策略
	PodType string `json:"pod_type"`

	PodImage string `json:"pod_image"`
}
