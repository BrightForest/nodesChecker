package main

import (
	"io/ioutil"
	"os"
	"strconv"
)

func getConfiguration() *Configuration{
	var config Configuration
	config.KubernetesApiAddr = os.Getenv("KUBERNETES_PORT_443_TCP_ADDR")
	config.ClusterAddr = "https://" + os.Getenv("KUBERNETES_PORT_443_TCP_ADDR") + os.Getenv("KUBERNETES_SERVICE_PORT_HTTP")
	config.AuthToken = os.Getenv("AUTH_TOKEN")
	config.ClusterApiConnTimeout, _ = strconv.Atoi(os.Getenv("API_CONN_TIMEOUT"))
	return &config
}

func init() {
	LogInit(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
}

func main() {
	curatorAddr := os.Getenv("CURATOR_ADDR")
	NodePod := podLoader(getConfiguration(), curatorAddr)
	Info.Println(NodePod.IP)
	Info.Println(NodePod.NodeName)
	StartWSServer(NodePod.IP)
}
