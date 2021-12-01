FROM alpine:3

COPY k8s-node-upgrade-agent /k8s-node-upgrade-agent

ENTRYPOINT ["/k8s-node-upgrade-agent"]
