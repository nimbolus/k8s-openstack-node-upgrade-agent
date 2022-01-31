package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/nimbolus/k8s-openstack-node-upgrade-agent/pkg/health"
	"github.com/nimbolus/k8s-openstack-node-upgrade-agent/pkg/openstack"
)

const (
	versionEnvVar   = "SYSTEM_UPGRADE_PLAN_LATEST_VERSION"
	imageNameEnvVar = "OPENSTACK_IMAGE_NAME"
)

func main() {
	verify := flag.Bool("verify", false, "verify cluster health for a given time period")
	duration := flag.Duration("duration", time.Minute, "duration for verify option")
	instanceUpgrade := flag.Bool("instanceUpgrade", false, "upgrades the k8s node instance")
	srvImageChannel := flag.Bool("serveImageChannel", true, "serve http endpoint for image channel")
	flag.Parse()

	if *verify {
		log.Printf("verifying cluster health for %s", duration.String())
		if err := health.VerifyClusterHealth(*duration); err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("cluster health verified")
	} else if *instanceUpgrade {
		latestImageName := os.Getenv(imageNameEnvVar)
		latestImageID := os.Getenv(versionEnvVar)

		if latestImageID == "" {
			latestImageID = "latest"
		}

		if err := openstack.UpdateInstanceImage(latestImageName, latestImageID); err != nil {
			log.Fatal(err.Error())
		}
	} else if *srvImageChannel {
		if err := openstack.ServeImageChannel(":8080"); err != nil {
			log.Fatalf("failed to serve image channel: %v", err)
		}
	} else {
		log.Fatal("no task given")
	}
}
