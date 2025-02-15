package glusterfs

import "github.com/pborman/uuid"

type Peer struct {
	PeerID    uuid.UUID
	BrickPath string
	BrickSize int64
}
