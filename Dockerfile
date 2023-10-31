FROM golang:1.19.0 AS builder

WORKDIR /go/src/github.com/sapcc/ccloud-nodeCIDR-controller
ADD . .
RUN CGO_ENABLED=0 go build -v -o /ccloud-nodeCIDR-controller .

FROM alpine:3.18.4
LABEL source_repository="https://github.com/sapcc/ccloud-nodeCIDR-controller"
COPY --from=builder /ccloud-nodeCIDR-controller /ccloud-nodeCIDR-controller
