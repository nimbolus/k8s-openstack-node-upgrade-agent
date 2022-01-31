FROM alpine:3

COPY k8s-openstack-node-upgrade-agent /k8s-openstack-node-upgrade-agent

ENTRYPOINT ["/k8s-openstack-node-upgrade-agent"]
