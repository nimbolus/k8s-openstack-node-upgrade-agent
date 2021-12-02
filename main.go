package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/nimbolus/k8s-node-upgrade-agent/pkg/health"
	"github.com/nimbolus/k8s-node-upgrade-agent/pkg/openstack"
)

const (
	versionEnvVar = "SYSTEM_UPGRADE_PLAN_LATEST_VERSION"
)

func main() {
	verify := flag.Bool("verify", false, "verify cluster health for a given time peroid")
	duration := flag.Duration("duration", time.Minute, "duration for verify option")
	instanceUpdate := flag.Bool("instanceUpdate", false, "updates the k8s node instance")
	srvImageChannel := flag.Bool("serveImageChannel", true, "serve http endpoint for image channel")
	flag.Parse()

	if *verify {
		log.Printf("verifying cluster health for %s", duration.String())
		if err := health.VerifyClusterHealth(*duration); err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("cluster health verified")
	} else if *instanceUpdate {
		latestImageID := os.Getenv(versionEnvVar)
		if latestImageID == "" {
			log.Fatalf("no latest image id given, please specify %s in environment", versionEnvVar)
		}

		if err := openstack.UpdateInstanceImage(latestImageID); err != nil {
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
