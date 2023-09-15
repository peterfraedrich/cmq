package main

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Cluster struct {
	ConfigMapName string
	Queues        map[string]*Queue
	k8sClient     *kubernetes.Clientset
	Config        *Config
}

func NewCluster(c *Config, k *kubernetes.Clientset) *Cluster {
	return &Cluster{
		ConfigMapName: c.Kube.ConfigMapName,
		Queues:        map[string]*Queue{},
		k8sClient:     k,
		Config:        c,
	}
}

func (c *Cluster) NewQueue(name string) error {
	q := &Queue{
		Name:   name,
		Index:  []string{},
		Length: 0,
		Client: c.k8sClient,
		Config: c.Config,
	}
	nameHash := HashString(name)
	c.Queues[nameHash] = q
	cmapName := fmt.Sprintf("%s-%s", q.Config.Kube.ConfigMapName, q.Name)
	ns := v1.Namespace{}
	ns.Name = q.Config.Kube.Namespace
	_, err := q.Client.CoreV1().Namespaces().Create(context.TODO(), &ns, metav1.CreateOptions{})
	if err != nil && err.Error() != fmt.Sprintf("namespaces \"%s\" already exists", q.Config.Kube.Namespace) {
		return err
	}
	cmap := v1.ConfigMap{}
	cmap.Name = cmapName
	b, err := json.Marshal([]string{})
	if err != nil {
		return err
	}
	cmap.BinaryData = map[string][]byte{}
	cmap.BinaryData["idx"] = b
	_, _ = q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).Create(context.TODO(), &cmap, metav1.CreateOptions{})
	return nil
}

func (c *Cluster) GetQueue(name string) (*Queue, error) {
	nameHash := HashString(name)
	if val, ok := c.Queues[nameHash]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("Name %s does not exist", name)
}

func HashString(name string) string {
	return HashBytes([]byte(name))
}

func HashBytes(b []byte) string {
	n := sha1.Sum(b)
	return string(n[:])
}

type Queue struct {
	Name   string
	Index  []string
	Length uint
	Client *kubernetes.Clientset
	Config *Config
}

func (q *Queue) GetQueueData() (idx []string, objects map[string][]byte, cm *v1.ConfigMap, err error) {
	cmi, err := q.Client.CoreV1().ConfigMaps(q.Config.Kube.ConfigMapName).Get(context.TODO(), fmt.Sprintf("%s-%s", q.Config.Kube.ConfigMapName, q.Name), metav1.GetOptions{})
	if err != nil {
		return nil, nil, cmi, err
	}
	// var idx []string <-- leave this for readability's sake
	err = json.Unmarshal(cmi.BinaryData["idx"], &idx)
	if err != nil {
		return nil, nil, cmi, err
	}
	return idx, cmi.BinaryData, cmi, nil
}

func (q *Queue) WriteQueueData(configMap *v1.ConfigMap, data map[string][]byte) error {
	configMap.BinaryData = data
	_, err := q.Client.CoreV1().ConfigMaps(q.Config.Kube.ConfigMapName).Update(context.TODO(), configMap, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (q *Queue) Push(i []byte) (string, error) {
	hash := HashBytes(i)
	idx, objects, cm, err := q.GetQueueData()
	if err != nil {
		return "", err
	}
	idx = append(idx, hash)
	idxb, err := json.Marshal(idx)
	if err != nil {
		return "", err
	}
	objects["idx"] = idxb
	objects[hash] = i
	err = q.WriteQueueData(cm, objects)
	if err != nil {
		return "", err
	}
	return hash, nil
}
