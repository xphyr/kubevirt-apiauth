package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/pflag"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubevirt.io/client-go/kubecli"
)

var flagNamespaceString string
var namespaces []string

func init() {
	pflag.StringVar(&flagNamespaceString, "namespaces", "", "additional namespaces to check")
}

func main() {

	// kubecli.DefaultClientConfig() prepares config using kubeconfig.
	// typically, you need to set env variable, KUBECONFIG=<path-to-kubeconfig>/.kubeconfig
	clientConfig := kubecli.DefaultClientConfig(&pflag.FlagSet{})

	// parse any command line flags we might have
	pflag.Parse()

	// retrive default namespace.
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		log.Fatalf("error in namespace : %v\n", err)
	}

	//add the default namespace to our slice of namespaces
	namespaces = append(namespaces, namespace)

	// check and see if we have any additional namespaces to check
	if flagNamespaceString != "" {
		// we have additional namespaces to check
		fmt.Println("additional namespaces to check are: ", flagNamespaceString)
		s := strings.Split(flagNamespaceString, ",")
		namespaces = append(namespaces, s...)
	}

	fmt.Println("Checking the following namespaces: ", namespaces)

	// get the kubevirt client, using which kubevirt resources can be managed.
	virtClient, err := kubecli.GetKubevirtClientFromClientConfig(clientConfig)
	if err != nil {
		log.Fatalf("cannot obtain KubeVirt client: %v\n", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 5, ' ', 0)
	fmt.Fprintln(w, "Type\tName\tNamespace\tStatus")

	//Now loop through all our namespaces

	for _, namespaceInstance := range namespaces {
		// Fetch list of VMs & VMIs
		vmList, err := virtClient.VirtualMachine(namespaceInstance).List(&k8smetav1.ListOptions{})
		if err != nil {
			log.Fatalf("cannot obtain KubeVirt vm list: %v\n", err)
		}
		vmiList, err := virtClient.VirtualMachineInstance(namespaceInstance).List(&k8smetav1.ListOptions{})
		if err != nil {
			log.Fatalf("cannot obtain KubeVirt vmi list: %v\n", err)
		}

		for _, vm := range vmList.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vm.Kind, vm.Name, vm.Namespace, vm.Status.Ready)
		}
		for _, vmi := range vmiList.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%v\n", vmi.Kind, vmi.Name, vmi.Namespace, vmi.Status.Phase)
		}
	}
	w.Flush()

	// this is a hack to allow the program to just run doing nothing until signaled to quit
	// this is used as an example to show what would happen if the application is run inside a k8s cluster
	if _, present := os.LookupEnv("KUBERNETES_SERVICE_HOST"); present {

		sigs := make(chan os.Signal, 1)
		done := make(chan bool, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigs
			fmt.Println()
			fmt.Println(sig)
			done <- true
		}()

		fmt.Println("awaiting signal")
		<-done
		fmt.Println("exiting")
	}
}
