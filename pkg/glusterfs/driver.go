package glusterfs

import (
	"net"
	"os"
	"sync"
	"time"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/gluster/gluster-csi-driver/pkg/glusterfs/config"
	"github.com/gluster/glusterd2/pkg/restclient"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"
)

// import (
// 	"time"

// 	"github.com/gluster/gluster-csi-driver/pkg/glusterfs/config"

// 	"github.com/gluster/glusterd2/pkg/restclient"
// 	"github.com/golang/glog"
// 	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
// )

const (
	glusterfsCSIDriverName    = "org.gluster.glusterfs"
	glusterfsCSIDriverVersion = "1.0.0"
)

// // GfDriver is the struct embedding information about the connection to gluster
// // cluster and configuration of CSI driver.
type GfDriver struct {
	client *restclient.Client
	*config.Config
}

// // New returns CSI driver
func New(config *config.Config) *GfDriver {
	gfd := &GfDriver{}

	if config == nil {
		glog.Errorf("GlusterFS CSI driver initialization failed")
		return nil
	}

	gfd.Config = config
	var err error
	gfd.client, err = restclient.NewClientWithOpts(
		restclient.WithBaseURL(config.RestURL),
		restclient.WithUsername(config.RestUser),
		restclient.WithPassword(config.RestSecret),
		restclient.WithTimeOut(time.Duration(config.RestTimeout)*time.Second),
		restclient.WithDebugRoundTripper())

	if err != nil {
		glog.Errorf("error creating glusterd2 REST client: %s", err.Error())
		return nil
	}

	glog.V(1).Infof("GlusterFS CSI driver initialized")

	return gfd
}

// NewControllerServer initialize a controller server for GlusterFS CSI driver.
func NewControllerServer(g *GfDriver) *ControllerServer {
	return &ControllerServer{
		GfDriver: g,
	}
}

// NewNodeServer initialize a node server for GlusterFS CSI driver.
func NewNodeServer(g *GfDriver) *NodeServer {
	return &NodeServer{
		GfDriver: g,
	}
}

// NewIdentityServer initialize an identity server for GlusterFS CSI driver.
func NewIdentityServer(g *GfDriver) *IdentityServer {
	return &IdentityServer{
		GfDriver: g,
	}
}

// // Run start a non-blocking grpc controller,node and identityserver for
// // GlusterFS CSI driver which can serve multiple parallel requests
func (g *GfDriver) Run() {
	// srv := csicommon.NewNonBlockingGRPCServer()
	srv := NewNonBlockingGRPCServer()
	srv.Start(g.Endpoint, NewIdentityServer(g), NewControllerServer(g), NewNodeServer(g), false)
	srv.Wait()
}

// Defines Non blocking GRPC server interfaces
type NonBlockingGRPCServer interface {
	// Start services at the endpoint
	Start(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, testMode bool)
	// Waits for the service to stop
	Wait()
	// Stops the service gracefully
	Stop()
	// Stops the service forcefully
	ForceStop()
}

func NewNonBlockingGRPCServer() NonBlockingGRPCServer {
	return &nonBlockingGRPCServer{}
}

// NonBlocking server
type nonBlockingGRPCServer struct {
	wg     sync.WaitGroup
	server *grpc.Server
}

func (s *nonBlockingGRPCServer) Start(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, testMode bool) {

	s.wg.Add(1)

	go s.serve(endpoint, ids, cs, ns, testMode)
}

func (s *nonBlockingGRPCServer) Wait() {
	s.wg.Wait()
}

func (s *nonBlockingGRPCServer) Stop() {
	s.server.GracefulStop()
}

func (s *nonBlockingGRPCServer) ForceStop() {
	s.server.Stop()
}

func (s *nonBlockingGRPCServer) serve(endpoint string, ids csi.IdentityServer, cs csi.ControllerServer, ns csi.NodeServer, testMode bool) {

	proto, addr, err := ParseEndpoint(endpoint)
	if err != nil {
		klog.Fatal(err.Error())
	}

	if proto == "unix" {
		addr = "/" + addr
		if err := os.Remove(addr); err != nil && !os.IsNotExist(err) {
			klog.Fatalf("Failed to remove %s, error: %s", addr, err.Error())
		}
	}

	listener, err := net.Listen(proto, addr)
	if err != nil {
		klog.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(logGRPC),
	}
	server := grpc.NewServer(opts...)
	s.server = server

	if ids != nil {
		csi.RegisterIdentityServer(server, ids)
	}
	if cs != nil {
		csi.RegisterControllerServer(server, cs)
	}
	if ns != nil {
		csi.RegisterNodeServer(server, ns)
	}

	// Used to stop the server while running tests
	if testMode {
		s.wg.Done()
		go func() {
			// make sure Serve() is called
			s.wg.Wait()
			time.Sleep(time.Millisecond * 1000)
			s.server.GracefulStop()
		}()
	}

	klog.Infof("Listening for connections on address: %#v", listener.Addr())

	err = server.Serve(listener)
	if err != nil {
		klog.Fatalf("Failed to serve grpc server: %v", err)
	}
}
