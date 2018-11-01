package main

import (
	"os"

	"github.com/tjamet/kubectl-switch/kubectl"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const dlTemplate = "https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl"

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func main() {
	configDir := homeDir() + "/.kube/"

	config, err := clientcmd.BuildConfigFromFlags("", configDir+"/config")
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	version, err := clientset.ServerVersion()
	if err != nil {
		panic(err)
	}
	if !kubectl.Installed(version.String()) {
		err := kubectl.Download(version.String())
		if err != nil {
			panic(err)
		}
	}
	os.Exit(kubectl.Exec(version.String(), os.Args[1:]...))
}
