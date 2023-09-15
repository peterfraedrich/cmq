package main

type kubeConfig struct {
	IsExternal     bool   `default:"false"`
	KubeconfigPath string `default:".kube/config"`
	Namespace      string `default:"cmq"`
	ConfigMapName  string `default:"default"`
	ShardSize      uint   `default:"10"`
	ShardPrefix    string `default:"cmq"`
}

type serverConfig struct {
	Host           string `default:"0.0.0.0"`
	Port           uint   `default:"5000"`
	UseSSL         bool   `default:"false"`
	CertKeyPath    string `default:"crt.key"`
	CertSecretPath string `default:"crt.secret"`
}

type uiConfig struct {
	Enabled bool `default:"true"`
	Port    uint `default:"8080"`
}

type Config struct {
	Debug  bool
	Kube   kubeConfig
	Server serverConfig
	UI     uiConfig
}
