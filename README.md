# k8s-node-upgrade-agent

Upgrade agent for OpenStack-based Kubernetes nodes, which use regular updated Glance images.
It uses [rancher/system-upgrade-controller](https://github.com/rancher/system-upgrade-controller) to drain the Kubernetes nodes one by one and rebuild the OpenStack instance with the latest image revision.

## Example

First create a secret named `openstack-clouds` with a `clouds.yaml` configuration file (see [OpenStack docs](https://docs.openstack.org/python-openstackclient/latest/cli/man/openstack.html#config-files)). Then use the following ressource to create an upgrade plan.

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
  # image ID (for testing, later a channel will be used)
  version: 5a095795-9015-499b-bb03-abf2cbc7e2ab
  drain:
    force: true
  prepare:
    image: registry.zotha.de/nimbolus/k8s-node-upgrade-agent:5d4d45be
    # verify cluster health for one minute before upgrading the next node
    args: ["-verify", "-duration=1m"]
  upgrade:
    image: registry.zotha.de/nimbolus/k8s-node-upgrade-agent:5d4d45be
```
