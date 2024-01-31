package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"

	timekeeper "github.com/andreistan26/TimeKeeper/proto"
	"github.com/go-redis/redis/v8"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"google.golang.org/grpc"
)

var (
	grpcPort = os.Getenv("TK_GRPC_PORT")

	tkDbPath = os.Getenv("TK_DB_PATH")

	influxDBToken  = os.Getenv("TK_INFLUXDB_TOKEN")
	influxDBURL    = "http://localhost:8086"
	influxDBOrg    = "TimeKeeper"
	influxDBBucket = "test"

	redisAddr = "http://localhost:6379"
	redisPass = os.Getenv("TK_REDIS_PASS")
)

type TimeKeeperServer struct {
	DB           *sql.DB
	InfluxClient influxdb2.Client
	RedisClient  *redis.Client
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

func (tk *TimeKeeperServer) ConnectRedis() error {
	tk.RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})
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
			return nil, err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		ret := tk.RedisClient.HSet(
			ctx,
			fmt.Sprint(id),
			map[string]interface{}{
				"machine": req.MachineName,
				"label":   "",
				"state":   "",
			},
		)
		if ret.Err() != nil {
			return nil, err
		}

		resp.Id = (uint64)(id)
	}
	return resp, nil
}

func (tk *TimeKeeperServer) SendData(ctx context.Context, req *timekeeper.SendDataRequest) (*timekeeper.SendDataResponse, error) {
	wapi := tk.InfluxClient.WriteAPI(influxDBOrg, influxDBBucket)
	var last_err error

	// check if point is the same as last point so we can skip it
	res, err := tk.RedisClient.HMGet(ctx, fmt.Sprint(req.GetId()), "machine", "label", "state").Result()
	if err != nil {
		last_err = err
		log.Printf("Error getting data from redis: %v", err)
	}
	machine, label, state := res[0], res[1], res[2]

	for _, data := range req.DataPoints {
		if data.GetLabel() == label && data.GetState().String() == state {
			continue
		}

		p := influxdb2.NewPoint(
			"test",
			map[string]string{"machine": machine.(string)},
			map[string]interface{}{"label": data.GetLabel()},
			data.GetTimestamp().AsTime(),
		)

		label = data.GetLabel()
		state = data.GetState().String()

		wapi.WritePoint(p)
	}

	ret := tk.RedisClient.HMSet(
		ctx,
		fmt.Sprint(req.Id),
		map[string]interface{}{
			"label": label,
			"state": state,
		},
	)

	if ret.Err() != nil {
		last_err = ret.Err()
		log.Printf("Error getting data from redis: %v", ret.Err())
	}

	wapi.Flush()
	return &timekeeper.SendDataResponse{}, last_err
}

func StartServer(ctx context.Context) error {
	// start db connection
	tkServer := &TimeKeeperServer{}
	tkServer.ConnectSqliteDB()
	defer tkServer.DB.Close()

	tkServer.ConnectInfluxDB()
	defer tkServer.InfluxClient.Close()

	tkServer.ConnectRedis()
	defer tkServer.RedisClient.Close()

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

	return nil
}