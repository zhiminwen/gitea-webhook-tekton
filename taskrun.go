package main

import (
	"fmt"
	"log"
	"time"

	"github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	tb "github.com/tektoncd/pipeline/test/builder"
	"k8s.io/client-go/rest"
)

func CreateTaskRun(namespace string, serviceAccount string, taskRefName string, gitSourceResourceName string, dockerImageResourceName string) error {
	// kubeconfig := flag.String("kubeconfig", "/Users/wenzm/.kube/config", "absolute path to the kubeconfig file")
	// flag.Parse()
	// config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	// if err != nil {
	// 	panic(err.Error())
	// }

	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	cs, err := versioned.NewForConfig(config)
	if err != nil {
		log.Printf("failed to create pipeline clientset: %s", err)
		return err
	}
	tRun := tb.TaskRun(
		fmt.Sprintf("taskrun-by-webhook-%d", time.Now().Unix()),
		namespace,
		tb.TaskRunLabel("create-by", "gitea-webhook"),
		tb.TaskRunSpec(
			tb.TaskRunServiceAccount(serviceAccount),
			tb.TaskRunTaskRef(taskRefName), // tb.TaskRefKind(v1alpha1.NamespacedTaskKind), tb.TaskRefAPIVersion("a1")),
			tb.TaskRunInputs(
				tb.TaskRunInputsResource("git-source", tb.TaskResourceBindingRef(gitSourceResourceName)),
			),
			tb.TaskRunOutputs(
				tb.TaskRunOutputsResource("docker-image", tb.TaskResourceBindingRef(dockerImageResourceName)),
			),
		),
	)

	tRun, err = cs.TektonV1alpha1().TaskRuns(namespace).Create(tRun)
	if err != nil {
		log.Printf("Failed to create TaskRun:%v", err)
		return err
	}
	return nil
}
