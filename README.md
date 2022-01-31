# k8s-openstack-node-upgrade-agent

Upgrade agent for OpenStack-based Kubernetes nodes, which are based on regular updated Glance images.
It uses [rancher/system-upgrade-controller](https://github.com/rancher/system-upgrade-controller) to drain the Kubernetes nodes one by one and rebuild the OpenStack instance with the latest image revision. It's also possible to migrate to another operating system by switching to a new image.

The upgrade process assumes that the OpenStack instances will provision itself via cloud-init after they got rebuilt. Also the cluster state should be stored elsewhere, e.g. in an OpenStack volume. Checkout [nimbolus/tf-k3s](https://github.com/nimbolus/tf-k3s) for such Kubernetes deployment.

## Usage

```sh
./k8s-openstack-node-upgrade-agent --help
  -duration duration
    	duration for verify option (default 1m0s)
  -instanceUpgrade
    	upgrades the k8s node instance
  -serveImageChannel
    	serve http endpoint for image channel (default true)
  -verify
    	verify cluster health for a given time period
```

## Deployment

First deploy the [rancher/system-upgrade-controller](https://github.com/rancher/system-upgrade-controller#deploying):
```sh
helm repo add nimbolus-k8s-openstack-node-upgrade-agent https://nimbolus.github.io/k8s-openstack-node-upgrade-agent
helm repo update
helm install -n system-upgrade --create-namespace system-upgrade-controller nimbolus-k8s-openstack-node-upgrade-agent/system-upgrade-controller
```

### Upgrade cluster to a given image version

After the system-upgrade-controller is ready, create a secret named `openstack-clouds` with OpenStack credentials like shown in [examples/openstack-clouds.yaml](./examples/openstack-clouds.yaml) (see also [OpenStack docs](https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html#config-files)).

Then checkout [examples/instance-upgrade-manual.yaml](./examples/instance-upgrade-manual.yaml) for an example upgrade plan.

### Use a channel for regular upgrades

On OpenStack clouds where images get rebuild regularly to include the latest kernels and security patches, the system-upgrade-controller needs a channel endpoint to check for new image IDs. The k8s-openstack-node-upgrade-agent can also fullfil this requirement by serving an HTTP endpoint, which returns the latest ID for a given OpenStack image name.

After the system-upgrade-controller is ready, deploy the node-upgrade-channel, remember to set the OpenStack credentials in the Helm values:
```sh
helm install -n system-upgrade -f value_overrides.yaml system-upgrade-controller nimbolus-k8s-openstack-node-upgrade-agent/node-upgrade-channel
```

Finally checkout [examples/instance-upgrade-channel.yaml](./examples/instance-upgrade-channel.yaml) for an example upgrade plan.
