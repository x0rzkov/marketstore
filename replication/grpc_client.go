package replication

import (
	"context"
	pb "github.com/alpacahq/marketstore/v4/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"io"
)

type GRPCReplicationClient struct {
	EnableSSL    bool
	Client       pb.ReplicationClient
	clientConn   *grpc.ClientConn
	streamClient pb.Replication_GetWALStreamClient
}

func NewGRPCReplicationClient(masterHost string, enableSSL bool) (*GRPCReplicationClient, error) {
	// TODO: implement SSL option
	conn, err := grpc.Dial(masterHost, grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect Master server")
	}

	c := pb.NewReplicationClient(conn)

	return &GRPCReplicationClient{
		EnableSSL:  enableSSL,
		Client:     c,
		clientConn: conn,
	}, nil
}

func (rc GRPCReplicationClient) Connect(ctx context.Context) (pb.Replication_GetWALStreamClient, error) {
	stream, err := rc.Client.GetWALStream(ctx, &pb.GetWALStreamRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wal message stream")
	}

	return stream, nil
}

func (rc GRPCReplicationClient) Recv() ([]byte, error) {
	resp, err := rc.streamClient.Recv()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to get wal message from gRPC stream")
	}
	if resp == nil {
		return nil, errors.New("nil message received from gRPC stream")
	}
	return resp.TransactionGroup, nil
}

func (rc GRPCReplicationClient) CloseSend() {
	rc.streamClient.CloseSend()
}

func (rc GRPCReplicationClient) Close() error {
	err := rc.streamClient.CloseSend()
	if err != nil {
		return errors.Wrap(err, "failed to close gRPC stream connection")
	}

	err = rc.clientConn.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close gRPC connection")
	}

	return nil
}