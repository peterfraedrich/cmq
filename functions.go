package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func Debug(c *Config, i interface{}) {
	if c.Debug {
		fmt.Printf(">>>>> DEBUG: %+v\n", i)
	}
}

func getFlags() *Config {
	debug := flag.Bool("debug", false, "Enable debug mode; more verbose logging and stuff")
	isExternal := flag.Bool("external", false, "(optional) use Kubeconfig if node is not running in a k8s cluster")
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	ns := flag.String("namespace", "cmq", "Namespace for CMQ to use")
	cmName := flag.String("name", "default", "Name of the primary config map to use")
	shardSize := flag.Uint("shardsize", 10, "Number of items in a shard")
	shardPrefix := flag.String("shardprefix", "cmq", "Prefix string for shards in this cluster")
	srvHost := flag.String("host", "0.0.0.0", "Host for API to listen on")
	srvPort := flag.Uint("port", 5000, "Port for API to listen on")
	ssl := flag.Bool("ssl", false, "Use SSL to secure API")
	certKey := flag.String("sslkey", "crt.key", "Path to certificate key")
	certSecret := flag.String("sslsecret", "crt.secret", "Path to certificate secret")
	useUI := flag.Bool("ui", true, "Enable web UI")
	uiPort := flag.Uint("uiport", 8080, "Port for the web UI to use")
	flag.Parse()
	return &Config{
		Debug: *debug,
		Kube: kubeConfig{
			IsExternal:     *isExternal,
			KubeconfigPath: *kubeconfig,
			Namespace:      *ns,
			ConfigMapName:  *cmName,
			ShardSize:      *shardSize,
			ShardPrefix:    *shardPrefix,
		},
		Server: serverConfig{
			Host:           *srvHost,
			Port:           *srvPort,
			UseSSL:         *ssl,
			CertKeyPath:    *certKey,
			CertSecretPath: *certSecret,
		},
		UI: uiConfig{
			Enabled: *useUI,
			Port:    *uiPort,
		},
	}

}

func buildKubeClient(c *Config) (*kubernetes.Clientset, error) {
	var clusterConfig *rest.Config
	var err error
	if !c.Kube.IsExternal {
		clusterConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		clusterConfig, err = clientcmd.BuildConfigFromFlags("", c.Kube.KubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(clusterConfig)
}

func prettyPrint(s interface{}) {
	j, _ := json.MarshalIndent(s, "", "    ")
	fmt.Print(string(j))
}
