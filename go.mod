module github.com/sapcc/ccloud-nodeCIDR-controller

replace github.com/hosting-de-labs/go-netbox => github.com/sapcc/go-netbox v0.0.0-20200225104756-2ba979226b5a

go 1.15

require (
	github.com/go-openapi/runtime v0.19.21
	github.com/hosting-de-labs/go-netbox v1.1.6-api2.7.6
	github.com/netbox-community/go-netbox v0.0.0-20200507032154-fbb6900a912a
	github.com/prometheus/client_golang v1.0.0
	go.uber.org/zap v1.16.0
	k8s.io/api v0.18.6
	k8s.io/client-go v0.18.6
	sigs.k8s.io/controller-runtime v0.6.3
)
