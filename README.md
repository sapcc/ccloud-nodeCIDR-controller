ccloud-nodeCIDR-controller
==========================

[![CI](https://github.com/sapcc/ccloud-nodeCIDR-controller/actions/workflows/ci.yaml/badge.svg)](https://github.com/sapcc/ccloud-nodeCIDR-controller/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/sapcc/ccloud-nodeCIDR-controller)](https://goreportcard.com/report/github.com/sapcc/ccloud-nodeCIDR-controller)

This repo contains a very simple controller that queries netbox for the nodeCIDR of a node.
If the node has no nodeCIDR set it is set to the value in netbox


Why?
----

The information is needed by Wormhole to create node specific tunnels.

