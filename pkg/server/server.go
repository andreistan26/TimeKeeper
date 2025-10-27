package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	timekeeper "github.com/andreistan26/TimeKeeper/pkg/gen/v1"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func getenvDefault(key, def string) string  {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

var (
	grpcPort = getenvDefault("TK_GRPC_PORT", "50051")
	httpPort = getenvDefault("TK_HTTP_PORT", "50050")

	// Clickhouse
	chHost = getenvDefault("TK_CH_HOST", "localhost")
	chPort = getenvDefault("TK_CH_PORT", "19000")
	chDatabase = getenvDefault("TK_CH_DATABASE", "timekeeper")
	chUser = getenvDefault("TK_CH_USER", "root")
	chPassword = getenvDefault("TK_CH_PASSWORD", "root")

)

type TimeKeeperServer struct {
	timekeeper.UnimplementedTimeKeeperServiceServer
	ChClient clickhouse.Conn
}

func NewTimeKeeperServer() (*TimeKeeperServer, error) {
	addr := fmt.Sprintf("%s:%s", chHost, chPort)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Username: chUser,
			Password: chPassword,
		},
		Debugf: func(format string, v ...any) {
			fmt.Printf(format, v...)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("could not connect to clickhouse %s: %v", addr, err)
	}

	return &TimeKeeperServer{
		ChClient: conn,
	}, nil
}

func (srv *TimeKeeperServer) EnsureDatabase(ctx context.Context) error {
	if err := srv.ChClient.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", chDatabase)); err != nil {
		return err
	}

	return srv.ChClient.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS timekeeper.data (
			timestamp DateTime64(9),
			machine String,
			tracker String,
			label String,
			state String
		) ENGINE = MergeTree
		  ORDER BY (timestamp)
	`)
}

func (tk *TimeKeeperServer) SendData(ctx context.Context, req *timekeeper.SendDataRequest) (*timekeeper.SendDataResponse, error) {
	if len(req.DataPoints) == 0 {
		return &timekeeper.SendDataResponse{}, nil
	}

 	batch, err := tk.ChClient.PrepareBatch(ctx, "INSERT INTO timekeeper.data (timestamp, machine, tracker, label, state)")
	if err != nil {
		return nil, err
	}

	for _, dp := range req.DataPoints {
		if err = batch.Append(
			dp.Timestamp.AsTime(),
			req.MachineName,
			req.TrackerName,
			dp.Label,
			dp.State,
		); err != nil {
			return &timekeeper.SendDataResponse{}, fmt.Errorf("batch append error: %v", err)
		}
	}

	if err := batch.Send(); err != nil {
		return &timekeeper.SendDataResponse{}, fmt.Errorf("batch send error: %v", err)
	}

	return &timekeeper.SendDataResponse{}, err
}

func StartGateway(ctx context.Context) error {
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%s", grpcPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("error while dialing: %v", err)
	}
	defer conn.Close()
	
	mux := runtime.NewServeMux()
	if err = timekeeper.RegisterTimeKeeperServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("failed to register the http server: %v", err)
	}

	addr := fmt.Sprintf("localhost:%s", httpPort)
	log.Println("HTTP API gateway server running on " + addr)
	if err = http.ListenAndServe(addr, mux); err != nil {
		return fmt.Errorf("gateway server closed: %v", err)
	}

	return nil
}

func StartServer(ctx context.Context) error {
	tkServer, err := NewTimeKeeperServer()
	if err != nil {
		return err
	}
	defer tkServer.ChClient.Close()

	if err = tkServer.EnsureDatabase(ctx); err != nil {
		return err
	}

	// start grpc server and sever data gatherers
	if grpcPort == "" {
		return fmt.Errorf("TK_GRPC_PORT env var not set")
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	timekeeper.RegisterTimeKeeperServiceServer(srv, tkServer)

	go func() {
		err := StartGateway(context.Background())
		if err != nil {
			log.Fatalf("gateway failed: %v", err)
		}
	}()

	log.Printf("TimeKeeper server started on port %s \n", grpcPort)
	if err := srv.Serve(listener); err != nil {
		return err
	}

	return nil
}
