package main

import (
	"flag"
	"github.com/xutao1989103/first-go-app/pkg/signals"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	masterUrl  string
	kubeconfig string
)

func main() {

	flag.Parse()

	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeconfig)
	if err != nil {

	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {

	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)

	controller := NewController(kubeClient, kubeInformerFactory.Apps().V1().Deployments())

	kubeInformerFactory.Start(stopCh)

	if err = controller.Run(stopCh); err != nil {

	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterUrl, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
