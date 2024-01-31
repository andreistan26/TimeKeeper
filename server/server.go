package server

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"

	timekeeper "github.com/andreistan26/TimeKeeper/proto"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"google.golang.org/grpc"
)

var (
	grpcPort       = os.Getenv("TK_GRPC_PORT")
	tkDbPath       = os.Getenv("TK_DB_PATH")
	influxDBToken  = os.Getenv("TK_INFLUXDB_TOKEN")
	influxDBURL    = "http://localhost:8086"
	influxDBOrg    = "TimeKeeper"
	influxDBBucket = "test"
)

type TimeKeeperServer struct {
	DB           *sql.DB
	InfluxClient influxdb2.Client
}

func (tk *TimeKeeperServer) ConnectSqliteDB() error {
	var err error
	if tkDbPath == "" {
		return errors.New("TK_DB_PATH env var not set")
	}
	if tk.DB, err = sql.Open("sqlite3", tkDbPath); err != nil {
		return err
	}

	_, err = tk.DB.Exec("CREATE TABLE IF NOT EXISTS data_sources (id INTEGER PRIMARY KEY AUTOINCREMENT, machine TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)")
	if err != nil {
		return err
	}
	return nil
}

func (tk *TimeKeeperServer) ConnectInfluxDB() error {
	tk.InfluxClient = influxdb2.NewClient(influxDBURL, influxDBToken)
	return nil
}

func (tk *TimeKeeperServer) Register(ctx context.Context, req *timekeeper.RegisterRequest) (*timekeeper.RegisterResponse, error) {
	resp := &timekeeper.RegisterResponse{}
	if req.Id == 0 {
		result, err := tk.DB.Exec("INSERT INTO data_sources (machine) VALUES (?)", req.MachineName)
		if err != nil {
			log.Printf("Error inserting data source: %v", err)
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting last inserted id: %v", err)
			return nil, err
		}

		resp.Id = (uint64)(id)
	}
	return resp, nil
}

func (tk *TimeKeeperServer) SendData(ctx context.Context, req *timekeeper.SendDataRequest) (*timekeeper.SendDataResponse, error) {
	wapi := tk.InfluxClient.WriteAPI(influxDBOrg, influxDBBucket)

	for _, data := range req.DataPoints {
		p := influxdb2.NewPoint(
			"test",
			map[string]string{"machine": data.}
		)
	}
	return &timekeeper.SendDataResponse{}, nil
}

func StartServer(ctx context.Context) error {
	// start db connection
	tkServer := &TimeKeeperServer{}
	tkServer.ConnectSqliteDB()
	defer tkServer.DB.Close()

	tkServer.ConnectInfluxDB()
	defer tkServer.InfluxClient.Close()

	// start grpc server and sever data gatherers
	if grpcPort == "" {
		return errors.New("TK_GRPC_PORT env var not set")
	}

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	srv := grpc.NewServer()
	timekeeper.RegisterTimeKeeperServiceServer(srv, tkServer)

	log.Println("TimeKeeper server started on port", grpcPort)
	if err := srv.Serve(listener); err != nil {
		return err
	}
}
