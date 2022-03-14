package gojenkins

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"strconv"
)

type KubernetesCloud struct {
	Raw      *[]CloudResponse
	K8sCloud CloudConfig
	Jenkins  *Jenkins
	Base     string
}

type CloudConfig struct {
	CloudName     string
	Namespace     string
	JenkinsURL    string
	JenkinsTunnel string
	Operation     string
}

type CloudResponse struct {
	Name                     string            `json:"name"`
	ServerUrl                string            `json:"serverUrl"`
	ServerCertificate        string            `json:"serverCertificate"`
	SkipTlsVerify            bool              `json:"skipTlsVerify"`
	JenkinsUrl               string            `json:"jenkinsUrl"`
	JenkinsTunnel            string            `json:"jenkinsTunnel"`
	Namespace                string            `json:"namespace"`
	AddMasterProxyEnvVars    bool              `json:"addMasterProxyEnvVars"`
	CapOnlyOnAlivePods       bool              `json:"capOnlyOnAlivePods"`
	UseJenkinsProxy          bool              `json:"useJenkinsProxy"`
	Labels                   map[string]string `json:"labels"`
	UsageRestricted          bool              `json:"usageRestricted"`
	PodRetention             string            `json:"podRetention"`
	ContainerCap             int               `json:"containerCap"`
	CredentialsID            string            `json:"credentialsID"`
	WaitForPodSec            int               `json:"waitForPodSec"`
	DirectConnection         bool              `json:"directConnection"`
	ReadTimeout              int               `json:"readTimeout"`
	MaxRequestsPerHost       int               `json:"maxRequestsPerHost"`
	DefaultsProviderTemplate string            `json:"defaultsProviderTemplate"`
	Templates                []string          `json:"templates"`
	PodLabels                map[string]string `json:"podLabels"`
	RetentionTimeout         int               `json:"retentionTimeout"`
	ConnectionTimeout        int               `json:"connectionTimeout"`
	WebSocket                bool              `json:"webSocket"`
}

var (
	temp *template.Template
	tpl  bytes.Buffer
)

func init() {
	temp = template.Must(template.ParseFiles("helpers/manage_clouds.tpl"))
}

func (k *KubernetesCloud) CloudConfigure(ctx context.Context) (*KubernetesCloud, error) {
	output, err := k.renderTemplate()
	if err != nil {
		return nil, err
	}
	data := map[string]string{
		"script": output,
	}
	r, err := k.Jenkins.Requester.Post(ctx, k.Base, nil, k.Raw, data)

	if err != nil {
		return nil, err
	}
	if r.StatusCode == 200 {
		return k, nil
	}

	return nil, errors.New(strconv.Itoa(r.StatusCode))
}

func (k *KubernetesCloud) renderTemplate() (string, error) {
	err := temp.Execute(&tpl, k.K8sCloud)
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}
