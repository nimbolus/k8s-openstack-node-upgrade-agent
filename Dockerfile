# build binary
FROM golang:1.17-alpine AS builder

COPY . /go/src/github.com/nimbolus/k8s-openstack-node-upgrade-agent

WORKDIR /go/src/github.com/nimbolus/k8s-openstack-node-upgrade-agent

RUN GOOS=linux CGO_ENABLED=0 go build -o k8s-openstack-node-upgrade-agent

# start clean for final image
FROM alpine:3

COPY --from=builder /go/src/github.com/nimbolus/k8s-openstack-node-upgrade-agent/k8s-openstack-node-upgrade-agent /k8s-openstack-node-upgrade-agent

ENTRYPOINT ["/k8s-openstack-node-upgrade-agent"]
