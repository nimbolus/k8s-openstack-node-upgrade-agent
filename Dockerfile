FROM alpine:3

COPY k8s-node-upgrade-agent /k8s-node-upgrade-agent

CMD ["/k8s-node-upgrade-agent"]
