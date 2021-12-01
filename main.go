package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

const (
	latestImageIDEnvVar = "SYSTEM_UPGRADE_PLAN_LATEST_VERSION"
	metadataUrl         = "http://169.254.169.254/openstack/latest/meta_data.json"
)

type metadata struct {
	UUID string `json="uuid"`
}

func main() {
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
