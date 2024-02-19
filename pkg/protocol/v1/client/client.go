package client

import (
	"context"
	"log"

	protocol_v1 "github.com/andreistan26/TimeKeeper/pkg/protocol/v1/protobuf"
	"google.golang.org/grpc"
)

type EventListener interface {
	StartEventStream() <-chan protocol_v1.DataPoint
}

type TKClient struct {
	eventListener EventListener
	grpcClient    protocol_v1.TimeKeeperServiceClient
	id            uint64
	conn          *grpc.ClientConn

	lastEvent protocol_v1.DataPoint
	dataQueue []*protocol_v1.DataPoint
}

type TimeKeeperRegistration struct {
	MachineName string
	TrackerName string
}

type TKClientOption func(*TKClient) error

func NewTKClient(ctx context.Context, conn *grpc.ClientConn, tkRegistration TimeKeeperRegistration, opts ...TKClientOption) (*TKClient, error) {
	tk := &TKClient{
		grpcClient: protocol_v1.NewTimeKeeperServiceClient(conn),
		conn:       conn,
		dataQueue:  make([]*protocol_v1.DataPoint, 1024),
	}

	resp, err := tk.grpcClient.Register(ctx, &protocol_v1.RegisterRequest{MachineName: tkRegistration.MachineName, TrackerName: tkRegistration.TrackerName})
	if err != nil {
		return nil, err
	}

	tk.id = resp.Id

	for _, opt := range opts {
		err := opt(tk)
		if err != nil {
			return nil, err
		}
	}

	return tk, nil
}

func (client *TKClient) Start(ctx context.Context) error {
	ch := client.eventListener.StartEventStream()
	for {
		select {
		case event := <-ch:
			if err := client.sendEvent(ctx, &event); err != nil {
				log.Println("Error sending event:", err)
				return err
			}
		}
	}
}

func (client *TKClient) sendEvent(ctx context.Context, event *protocol_v1.DataPoint) error {
	if event.Label == client.lastEvent.Label {
		return nil
	}

	if len(client.dataQueue) < 1024 {
		client.dataQueue = append(client.dataQueue, event)
	} else {
		log.Println("Data queue is full, dropping data")
	}

	_, err := client.grpcClient.SendData(ctx, &protocol_v1.SendDataRequest{
		Id:         client.id,
		DataPoints: client.dataQueue,
	})

	if err != nil {
		return err
	} else {
		client.lastEvent = *event
		client.dataQueue = client.dataQueue[:0]
	}
	return nil
}

func WithEventListener(el EventListener) TKClientOption {
	return func(tk *TKClient) error {
		tk.eventListener = el
		return nil
	}
}
