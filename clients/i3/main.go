package main

import (
	"context"
	"log"
	"os"
	"time"

	timekeeper "github.com/andreistan26/TimeKeeper/proto"
	"go.i3wm.org/i3/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:  "tk-i3",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SetContext(context.Background())
		return nil
	},
}

type ConnectionState int

const (
	CONN_ACTIVE ConnectionState = iota
	CONN_RECONNECTING
	CONN_DISCONNECTED
)

var startCommand = &cobra.Command{
	Use:  "start",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		client := &TKClient{
			EventChannel: StartWindowEventListener(),
			GrpcClient: timekeeper.NewTimeKeeperServiceClient(conn),
		}
		defer close(client.EventChannel)
		client.State = CONN_ACTIVE

		resp, err := client.GrpcClient.Register(ctx, &timekeeper.RegisterRequest{
			MachineName: "desktop",
		})
		if err != nil {
			return err
		}

		client.Id = resp.Id

		client.StartDataDispatcher(ctx)

		return nil
	},
}

var (
	LastDataPoint = timekeeper.DataPoint{}
	DataPointQueue []*timekeeper.DataPoint = make([]*timekeeper.DataPoint, 0) 
)

type TKClient struct {
	EventChannel chan i3.WindowEvent
	GrpcClient   timekeeper.TimeKeeperServiceClient
	Id 		 	 uint64
	State 	  	 ConnectionState
}



func StartWindowEventListener() chan i3.WindowEvent {
	recv := i3.Subscribe(i3.WindowEventType)
	c := make(chan i3.WindowEvent, 16)
	go func(recv *i3.EventReceiver, c chan i3.WindowEvent) {
		for recv.Next() {
			c <- *(recv.Event().(*i3.WindowEvent))
		}
	} (recv, c);
	return c;
}

func (client *TKClient ) PushDataPoint(ctx context.Context, event *i3.WindowEvent) error {
	// skip if same label as last point
	if event.Container.WindowProperties.Class == LastDataPoint.Label {
		return nil
	}

	dataPoint := &timekeeper.DataPoint{
		Timestamp: timestamppb.Now(),
		Label: event.Container.WindowProperties.Class,
		State: timekeeper.DataPointState_SAMPLE,
	}

	if len(DataPointQueue) < 1024 {
		DataPointQueue = append(DataPointQueue, dataPoint)
	}

	//if client.State == CONN_ACTIVE {
		_, err := client.GrpcClient.SendData(ctx, &timekeeper.SendDataRequest{
			Id: client.Id,
			DataPoints: DataPointQueue,
		})
		if err != nil {
			//client.State = CONN_DISCONNECTED
			if len(DataPointQueue) == 1024 {
				log.Println("DataPointQueue is full, dropping data point")
			}
			return err
		} else {
			LastDataPoint = *dataPoint
			DataPointQueue = DataPointQueue[:0]
		}
	//}

	return nil
}

func (client *TKClient) StartDataDispatcher(ctx context.Context) {
	for {
		select {
		case event := <-client.EventChannel:
			if err := client.PushDataPoint(ctx, &event); err != nil {
				log.Println(err)
				//client.State = CONN_RECONNECTING
				//go client.Recconect(ctx)
			}
		}
	}
}

func (client *TKClient) Recconect(ctx context.Context) {
	for {
		conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		client.GrpcClient = timekeeper.NewTimeKeeperServiceClient(conn)
		//client.State = CONN_ACTIVE
		return
	}
}

func main() {
	rootCommand.AddCommand(startCommand)
	if err := rootCommand.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
