ccloud-nodeCIDR-controller
==========================

This repo contains a very simple controller that queries netbox for the nodeCIDR of a node.
If the node has no nodeCIDR set it is set to the value in netbox


Why?
----

The information is needed by Wormhole to create node specific tunnels.

