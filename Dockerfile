FROM golang:1.21.3 AS builder

WORKDIR /go/src/github.com/sapcc/ccloud-nodeCIDR-controller
ADD . .
RUN CGO_ENABLED=0 go build -v -o /ccloud-nodeCIDR-controller .

FROM alpine
LABEL source_repository="https://github.com/sapcc/ccloud-nodeCIDR-controller"
COPY --from=builder /ccloud-nodeCIDR-controller /ccloud-nodeCIDR-controller
