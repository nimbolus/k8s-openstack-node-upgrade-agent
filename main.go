package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/openstack/clientconfig"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	latestImageIDEnvVar = "SYSTEM_UPGRADE_PLAN_LATEST_VERSION"
	metadataUrl         = "http://169.254.169.254/openstack/latest/meta_data.json"
	waitCheckInterval   = 10 * time.Second
	waitTimeout         = time.Hour
)

type metadata struct {
	UUID string `json="uuid"`
}

func isReady(n *v1.Node) bool {
	var cond v1.NodeCondition
	for _, c := range n.Status.Conditions {
		if c.Type == v1.NodeReady {
			cond = c
		}
	}
	return cond.Status == v1.ConditionTrue
}

func verifyClusterHealth(d time.Duration) (err error) {
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

func main() {
	verify := flag.Bool("verify", false, "verify cluster health for a given time peroid")
	duration := flag.Duration("duration", time.Minute, "duration for verify option")
	flag.Parse()

	if *verify {
		log.Printf("verifying cluster health for %s", duration.String())
		if err := verifyClusterHealth(*duration); err != nil {
			log.Fatal(err)
		}
		log.Printf("cluster health verified")
		os.Exit(0)
	}

	latestImageID := os.Getenv(latestImageIDEnvVar)
	if latestImageID == "" {
		log.Fatalf("no latest image id given, please specify %s in environment", latestImageIDEnvVar)
	}

	res, err := http.Get(metadataUrl)
	if err != nil {
		log.Fatalf("failed to get instance metadata: %v", err)
	}

	defer res.Body.Close()

	md, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("failed to read instance metadata: %v", err)
	}

	var m metadata
	if err = json.Unmarshal(md, &m); err != nil {
		log.Fatalf("failed to unmarshal instance metadata: %v", err)
	}

	log.Printf("instance id: %s", m.UUID)

	opts := &clientconfig.ClientOpts{
		Cloud: "openstack",
	}

	compute, err := clientconfig.NewServiceClient("compute", opts)
	if err != nil {
		log.Fatalf("failed to initialize openstack compute client: %v", err)
	}

	image, err := images.Get(compute, latestImageID).Extract()
	if err != nil {
		log.Fatalf("failed to fetch image: %v", err)
	}

	log.Printf("latest image: %s (%s)", image.ID, image.Name)

	server, err := servers.Get(compute, m.UUID).Extract()
	if err != nil {
		log.Fatalf("failed to fetch instance: %v", err)
	}

	if _, ok := server.Image["id"]; !ok {
		log.Fatalf("instance has no image id attribute")
	}

	if image.ID != server.Image["id"] {
		log.Printf("instance needs to be upgraded from image id %s to %ss", server.Image["id"], image.ID)

		rebuildOpts := servers.RebuildOpts{
			ImageRef: image.ID,
		}
		res := servers.Rebuild(compute, server.ID, rebuildOpts)
		if res.Err != nil {
			log.Fatalf("failed to rebuild instance: %v", res.Err)
		}
		log.Print("rebuild started ...")
	} else {
		log.Print("instance is up-to-date")
	}
}
