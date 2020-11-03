package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sapcc/go-netbox-go/dcim"
	"github.com/sapcc/go-netbox-go/ipam"
	"github.com/sapcc/go-netbox-go/models"
	uberzap "go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	net2 "net"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	netboxFails = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "netbox_fails",
				Help: "number of failed netbox requests",
			},
			)
	netboxResultFails = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "netbox_result_fails",
				Help: "number of times netbox results are too few or too many",
			},
		)
	k8sFails = prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "k8s_fails",
				Help: "number of times k8s operations failed",
			},
		)
)

func init() {
	metrics.Registry.MustRegister(netboxFails, netboxResultFails, k8sFails)
}

func main() {
	var debug bool
	var kubeContext string
	var netboxUrl string
	var netboxToken string
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.StringVar(&kubeContext, "kubecontext", "", "the context to use from kube_config")
	flag.StringVar(&netboxUrl, "netboxurl", "https://netbox.global.cloud.sap", "the netbox to query for the node")
	flag.StringVar(&netboxToken, "netboxtoken", "", "the netbox token to use")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
		if ! debug {
			o.Level = uberzap.NewAtomicLevelAt(uberzap.InfoLevel)
		}
	}))
	var opts manager.Options
	opts.MetricsBindAddress = ":8080"
	opts.HealthProbeBindAddress = ":8081"

	var log = logf.Log.WithName("ccloud-nodeCIDR-controller")

	if kubeContext == "" {
		kubeContext = os.Getenv("KUBECONTEXT")
	}
	restConfig, err := config.GetConfigWithContext(kubeContext)
	if err != nil {
		log.Error(err, "unable to setup config")
		os.Exit(1)
	}
	mgr, err := manager.New(restConfig, opts)
	if err != nil {
		log.Error(err, "cloud not create manager")
		os.Exit(1)
	}
	c, err := controller.New("ccloud-nodeCIDR-controller", mgr, controller.Options{
		Reconciler: reconcile.Func(func(request reconcile.Request) (reconcile.Result, error) {
			log.Info(fmt.Sprintf("Node: %s", request.Name))
			node := &corev1.Node{}
			mgr.GetClient().Get(context.Background(), request.NamespacedName, node)
			if node.Spec.PodCIDR == "" {
				log.Info("No PodCIDR set getting from netbox")
				nbIpam, err := ipam.New(netboxUrl, netboxToken)
				if err != nil {
					return reconcile.Result{}, err
				}
				opts := models.ListIpAddressesRequest{}
				opts.Q = node.Name
				res, err := nbIpam.ListIpAddresses(opts)
				if err != nil {
					log.Error(err, "error searching ips for hostname")
					netboxFails.Inc()
					return reconcile.Result{}, err
				}
				if res.Count != 1 {
					err := fmt.Errorf("too many results: got %d results for %s", res.Count, node.Name)
					log.Error(err, "error getting node ip")
					netboxResultFails.Inc()
					return reconcile.Result{},err
				}
				log.Info(fmt.Sprintf("%+v", res.Results[0]))
				deviceID := res.Results[0].AssignedInterface.Device.Id
				var interfaceName string
				interfaceName = "cbr0"
				log.Info(fmt.Sprintf("looking for device %d and interface %s", deviceID, interfaceName))
				//deviceParams := dcim.NewDcimInterfacesListParams().WithDeviceID(&deviceID).WithName(&interfaceName)
				interfaceOpts := models.ListInterfacesRequest{
					DeviceId: deviceID,
				}
				interfaceOpts.Name = interfaceName
				//cbr0, err := nb.Dcim.DcimInterfacesList(deviceParams, nil)
				nbDcim, err := dcim.New(netboxUrl, netboxToken)
				if err != nil {
					log.Error(err, "error getting dcim client")
					netboxFails.Inc()
					return reconcile.Result{}, err
				}
				interf, err := nbDcim.ListInterfaces(interfaceOpts)
				if err != nil {
					log.Error(err, "error searching interfaces")
					netboxFails.Inc()
					return reconcile.Result{}, err
				}
				if interf.Count != 1 {
					err := fmt.Errorf("too many results: got %d results for device %d", interf.Count, deviceID)
					log.Error(err, "error getting node device")
					netboxResultFails.Inc()
					return reconcile.Result{},err
				}
				//ipParams := ipam.NewIPAMIPAddressesListParams().WithInterfaceID(&cbr0.Payload.Results[0].ID)
				ipOpts := models.ListIpAddressesRequest{}
				ipOpts.InterfaceId = interf.Results[0].Id
				//theIP, err := nb.IPAM.IPAMIPAddressesList(ipParams, nil)
				theIP, err := nbIpam.ListIpAddresses(ipOpts)
				if err != nil {
					log.Error(err, "error searching cbr0 ip")
					netboxFails.Inc()
					return reconcile.Result{}, err
				}
				if theIP.Count != 1 {
					err := fmt.Errorf("too many results: got %d results for interface %d", theIP.Count, interf.Results[0].Id)
					log.Error(err, "error getting node device")
					netboxResultFails.Inc()
					return reconcile.Result{},err
				}
				_, net, err := net2.ParseCIDR(theIP.Results[0].Address)
				log.Info(fmt.Sprintf("net: %s", net.String()))
				node.Spec.PodCIDR = net.String()
				err = mgr.GetClient().Update(context.Background(), node)
				if err != nil {
					log.Error(err, "error updating node")
					k8sFails.Inc()
					return reconcile.Result{}, err
				}
			}
			return reconcile.Result{}, nil
		}),
	})
	if err != nil {
		log.Error(err, "unable to create ccloud-nodeCIDR-controller")
		os.Exit(1)
	}
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "unable to watch nodes")
		k8sFails.Inc()
		os.Exit(1)
	}
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "unable to continue running manager")
		os.Exit(1)
	}
}
