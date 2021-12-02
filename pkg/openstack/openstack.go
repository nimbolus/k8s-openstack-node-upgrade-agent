package openstack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/images"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	glanceImages "github.com/gophercloud/gophercloud/openstack/imageservice/v2/images"
	"github.com/gophercloud/utils/openstack/clientconfig"
	"github.com/gorilla/mux"
)

const (
	metadataUrl = "http://169.254.169.254/openstack/latest/meta_data.json"
	cloudName   = "openstack"
)

type metadata struct {
	UUID string `json="uuid"`
}

func getInstanceID() (string, error) {
	res, err := http.Get(metadataUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get instance metadata: %v", err)
	}

	defer res.Body.Close()

	md, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read instance metadata: %v", err)
	}

	var m metadata
	if err = json.Unmarshal(md, &m); err != nil {
		return "", fmt.Errorf("failed to unmarshal instance metadata: %v", err)
	}

	return m.UUID, nil
}

func getClient(service string) (*gophercloud.ServiceClient, error) {
	opts := &clientconfig.ClientOpts{
		Cloud: cloudName,
	}

	compute, err := clientconfig.NewServiceClient(service, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize openstack compute client: %v", err)
	}

	return compute, nil
}

func UpdateInstanceImage(imageName, latestImageID string) error {
	instanceID, err := getInstanceID()
	if err != nil {
		return err
	}

	log.Printf("instance id: %s", instanceID)

	compute, err := getClient("compute")
	if err != nil {
		return err
	}

	if latestImageID == "latest" {
		if imageName == "" {
			return fmt.Errorf("no image id nor name given")
		}
		latestImageID, err = getLatestImageID(imageName)
		if err != nil {
			return fmt.Errorf("failed to fetch latest image id: %v", err)
		}
	}

	image, err := images.Get(compute, latestImageID).Extract()
	if err != nil {
		return fmt.Errorf("failed to fetch image: %v", err)
	}

	log.Printf("latest image: %s (%s)", image.ID, image.Name)

	server, err := servers.Get(compute, instanceID).Extract()
	if err != nil {
		return fmt.Errorf("failed to fetch instance: %v", err)
	}

	if _, ok := server.Image["id"]; !ok {
		return fmt.Errorf("instance has no image id attribute")
	}

	if image.ID != server.Image["id"] {
		log.Printf("instance needs to be upgraded from image id %s to %s", server.Image["id"], image.ID)

		rebuildOpts := servers.RebuildOpts{
			ImageRef: image.ID,
		}
		res := servers.Rebuild(compute, server.ID, rebuildOpts)
		if res.Err != nil {
			return fmt.Errorf("failed to rebuild instance: %v", res.Err)
		}
		log.Print("rebuild started ...")
	} else {
		log.Print("instance is up-to-date")
	}
	return nil
}

func getLatestImageID(name string) (string, error) {
	glance, err := getClient("image")
	if err != nil {
		return "", fmt.Errorf("failed to initialize openstack glance client: %v", err)
	}

	opts := glanceImages.ListOpts{
		Name:   name,
		Status: "active",
		Sort:   "created_at:desc",
	}

	allPages, err := glanceImages.List(glance, opts).AllPages()
	if err != nil {
		return "", fmt.Errorf("failed to list glance images: %v", err)
	}

	allImages, err := glanceImages.ExtractImages(allPages)
	if err != nil {
		return "", fmt.Errorf("failed to extract glance images: %v", err)
	}

	if len(allImages) > 0 {
		return allImages[0].ID, nil
	} else {
		return "", nil
	}
}

func ServeImageChannel(addr string) error {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/openstack/images/{name}/latest", imageChannelHander)

	log.Printf("serving image channel on %s", addr)
	return http.ListenAndServe(addr, router)
}

func imageChannelHander(w http.ResponseWriter, r *http.Request) {
	imageName := mux.Vars(r)["name"]

	log.Printf("%s requests latest image id for %s", r.RemoteAddr, imageName)

	imageID, err := getLatestImageID(imageName)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "failed to fetch latest image id", http.StatusInternalServerError)
	} else if imageID == "" {
		msg := fmt.Sprintf("no image with name %s found", imageName)
		log.Print(msg)
		http.Error(w, msg, http.StatusNotFound)
	} else {
		log.Printf("latest image id for %s is %s", imageName, imageID)
		location := fmt.Sprintf("http://%s/openstack/images/%s/%s", r.Host, imageName, imageID)
		http.Redirect(w, r, location, http.StatusTemporaryRedirect)
	}
}
