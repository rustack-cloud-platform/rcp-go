package rustack

import (
	"net/url"
)

type Kubernetes struct {
	manager *Manager
	ID      string `json:"id"`
	Name    string `json:"name"`
	Locked  bool   `json:"locked"`
	Vdc     *Vdc   `json:"vdc"`

	Project struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`

	Floating     *Port  `json:"floating"`
	JobId        string `json:"job_id"`
	NodeCpu      int    `json:"node_cpu"`
	NodeDiskSize int    `json:"node_disk_size"`

	NodeRam            float64         `json:"node_ram"`
	NodeStorageProfile *StorageProfile `json:"node_storage_profile"`
	NodesCount         int             `json:"nodes_count"`
	Template           *Template       `json:"template"`
	UserPublicKey      string          `json:"user_public_key"`
}

func NewKubernetes(name string, vdc *Vdc) Kubernetes {
	v := Kubernetes{Name: name, Vdc: vdc}
	return v
}

func (m *Manager) GetKubernetes(id string) (k8s Kubernetes, err error) {
	path, _ := url.JoinPath("v1/kubernetes", id)
	err = m.Get(path, Defaults(), &k8s)
	if err != nil {
		return
	}
	k8s.Vdc.manager = m
	return
}
