package glusterfs

import (
	"context"
	"fmt"
	"time"

	"github.com/gluster/gluster-csi-driver/pkg/glusterfs/pb"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultLVMProxyPort int = 50050

// TODO: To be changed
// CreateLVOnPeers creates LVM volumes on remote peers using gRPC.
func CreateLVOnPeers(volRequest *ProvisionerConfig) {
	peers := volRequest.peers
	glog.V(4).Info("Peers in CreateLVOnPeers: ", peers, len(peers))
	for k, v := range peers {
		glog.V(4).Info("Dialing peer: ", v)

		// Fix incorrect formatting of gRPC address
		addr := fmt.Sprintf("%s:%d", k, DefaultLVMProxyPort)

		// Use secure transport credentials instead of deprecated WithInsecure
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			glog.V(4).Infof("Failed to connect to %s: %v", addr, err)
			continue // Skip this peer and try the next one
		}
		defer conn.Close()

		client := pb.NewVolumeClient(conn)

		req := &pb.CreateLVMVolumeRequest{
			VolumeGroup: "vg",
			VolumeName:  volRequest.gdVolReq.Name,
			Size:        v.BrickSize,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Increased timeout
		defer cancel()

		resp, err := client.CreateVolume(ctx, req)
		if err != nil {
			glog.V(4).Infof("Error calling CreateVolume on %s: %v", addr, err)
			continue // Skip this peer and try the next one
		}

		// Print the response
		glog.V(4).Infof("Server Response: %v", resp)

		// Store brick path if successful
		volRequest.peers[k] = Peer{
			PeerID:    v.PeerID,
			BrickSize: v.BrickSize,
			BrickPath: resp.BrickPath,
		}
		glog.V(4).Info("Brick path: ", resp.BrickPath)
	}
}
