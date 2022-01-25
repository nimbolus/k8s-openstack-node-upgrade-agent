# k8s-node-upgrade-agent

Upgrade agent for OpenStack-based Kubernetes nodes, which use regular updated Glance images.
It uses [rancher/system-upgrade-controller](https://github.com/rancher/system-upgrade-controller) to drain the Kubernetes nodes one by one and rebuild the OpenStack instance with the latest image revision.

## Example

First create a secret named `openstack-clouds` with a `clouds.yaml` configuration file (see [OpenStack docs](https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html#config-files)). Then use the following resource to create an upgrade plan.

Upgrade plan:
```yaml
apiVersion: upgrade.cattle.io/v1
kind: Plan
metadata:
  name: openstack-instance-plan
  namespace: system-upgrade
spec:
  concurrency: 1
  nodeSelector:
    matchExpressions:
      - key: plan.upgrade.cattle.io/openstack-instance
        operator: Exists
  tolerations:
    - key: node-role.kubernetes.io/master
      operator: Exists
  serviceAccountName: system-upgrade
  secrets:
    # clouds.yaml secret with OpenStack credentials
    - name: openstack-clouds
      path: /etc/openstack
  channel: http://k8s-node-upgrade-channel:8080/openstack/images/ubuntu-20.04/latest
  # instead of a channel also an image ID can be used by setting the `version` attribute
  # version: 5a095795-9015-499b-bb03-abf2cbc7e2ab
  drain:
    force: true
  prepare:
    image: registry.zotha.de/nimbolus/k8s-node-upgrade-agent:v0.2.0
    # verify cluster health for one minute before upgrading the next node
    args: ["-verify", "-duration=1m"]
  upgrade:
    image: registry.zotha.de/nimbolus/k8s-node-upgrade-agent:v0.2.0
    args: ["-instanceUpgrade"]
    envs:
      - name: OPENSTACK_IMAGE_NAME
        value: ubuntu-20.04
```

When an upgrade channel is used checkout `./deploy/upgrade-channel-manifest.yml` for how to deploy a channel service.

## Add Helm repo

```sh
helm repo add nimbolus https://nimbolus.pages.zotha.de/k8s-node-upgrade-agent
helm repo update
```
