package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"

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
	cmap.Data = map[string]string{}
	cmap.Data["idx"] = string(b[:])
	_, err = q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).Create(context.TODO(), &cmap, metav1.CreateOptions{})
	if err != nil && err.Error() != fmt.Sprintf("configmaps \"%s\" already exists", cmapName) {
		return err
	}
	return nil
}

func (c *Cluster) GetQueue(name string) (*Queue, error) {
	nameHash := HashString(name)
	if val, ok := c.Queues[nameHash]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("name %s does not exist", name)
}

func HashString(name string) string {
	return HashBytes([]byte(name))
}

func HashBytes(b []byte) string {
	h := sha1.New()
	h.Write(b)
	s := hex.EncodeToString(h.Sum(nil))
	return s
}

type Queue struct {
	Name   string
	Index  []string
	Length int
	Client *kubernetes.Clientset
	Config *Config
}

func (q *Queue) GetQueueData() (idx []string, objects map[string]string, cm *v1.ConfigMap, err error) {
	cmi, err := q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).Get(context.TODO(), fmt.Sprintf("%s-%s", q.Config.Kube.ConfigMapName, q.Name), metav1.GetOptions{})
	if err != nil {
		return nil, nil, cmi, err
	}
	// var idx []string <-- leave this for readability's sake
	err = json.Unmarshal([]byte(cmi.Data["idx"]), &idx)
	if err != nil {
		return nil, nil, cmi, err
	}
	return idx, cmi.Data, cmi, nil
}

func (q *Queue) WriteQueueData(configMap *v1.ConfigMap, idx []string, data map[string]string) error {
	idxb, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	data["idx"] = string(idxb[:])
	configMap.Data = data
	cmList, err := q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	exists := false
	for item := range cmList.Items {
		if cmList.Items[item].Name == fmt.Sprintf("%s-%s", q.Config.Kube.ConfigMapName, q.Name) {
			exists = true
		}
	}
	if !exists {
		configMap.ResourceVersion = ""
		_, err = q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).Create(context.TODO(), configMap, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		_, err = q.Client.CoreV1().ConfigMaps(q.Config.Kube.Namespace).Update(context.TODO(), configMap, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func (q *Queue) Push(i []byte) (string, error) {
	hash := HashBytes(i)
	idx, objects, cm, err := q.GetQueueData()
	if err != nil {
		return "", err
	}
	if !slices.Contains(idx, hash) {
		idx = append(idx, hash)
	}
	objects[hash] = string(i[:])
	err = q.WriteQueueData(cm, idx, objects)
	if err != nil {
		return "", err
	}
	q.Length = len(idx)
	return hash, nil
}

func (q *Queue) Pop() ([]byte, error) {
	idx, objects, cm, err := q.GetQueueData()
	if err != nil {
		return nil, err
	}
	if len(idx) == 0 {
		return nil, fmt.Errorf("no items in queue to pop")
	}
	i := idx[0]
	o := objects[i]
	idx = append(idx[:0], idx[0+1:]...)
	delete(objects, i)
	err = q.WriteQueueData(cm, idx, objects)
	if err != nil {
		return nil, err
	}
	q.Length = len(idx)
	return []byte(o), nil
}
