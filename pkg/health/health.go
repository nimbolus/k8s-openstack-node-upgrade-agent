package health

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	waitCheckInterval = 10 * time.Second
	waitTimeout       = time.Hour
)

func isReady(n *v1.Node) bool {
	var cond v1.NodeCondition
	for _, c := range n.Status.Conditions {
		if c.Type == v1.NodeReady {
			cond = c
		}
	}
	return cond.Status == v1.ConditionTrue
}

func VerifyClusterHealth(d time.Duration) (err error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("failed to get kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("failed to initialize kubernetes client: %v", err)
	}

	attempts := d.Seconds() / waitCheckInterval.Seconds()
	start := time.Now()
	for i := attempts; i > 0; i-- {
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("failed to list nodes, retrying: %v", err)
			i = attempts + 1
		} else {
			for _, n := range nodes.Items {
				if !isReady(&n) {
					log.Printf("node %s is not ready, reseting interval", n.Name)
					i = attempts + 1
					break
				}
			}
		}

		if time.Since(start) > waitTimeout {
			return fmt.Errorf("verify timeout of %s exceeded", waitTimeout.String())
		}

		if i <= attempts {
			log.Printf("cluster is healthy, checking again in %s", waitCheckInterval.String())
		}
		time.Sleep(waitCheckInterval)
	}

	return nil
}
