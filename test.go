package main

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")

	k8s, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(k8s)
	if err != nil {
		panic(err)
	}

	version, _ := clientset.Discovery().ServerVersion()
	fmt.Println(version)
	fmt.Println("===")

	apiList, _ := clientset.Discovery().ServerGroups()
	for _, api := range apiList.Groups {
		fmt.Printf("%s : %s \n", api.Name, api.Versions[0].Version)
	}
	fmt.Println("===")

	resourceList, _ := clientset.Discovery().ServerResources()
	for _, r := range resourceList {
		fmt.Printf("%s : %s \n", r.GroupVersion, r.APIResources[0].Name)
	}
}