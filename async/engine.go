package async

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/wayt/async/pb/server"
	pbWorker "github.com/wayt/async/pb/worker"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

var (
	config        = viper.New()
	DefaultEngine *Engine

	ErrCannotConnect = errors.New("cannot connect to server")
)

func init() {
	config.SetEnvPrefix("async")
	hostname, _ := os.Hostname()

	config.SetDefault("id", hostname)
	// config.SetDefault("bind", ":8179")
	config.SetDefault("advertise_addr", "127.0.0.1:8179")
	config.SetDefault("server_addr", "127.0.0.1:8080")
	config.SetDefault("worker", 2)

	config.AutomaticEnv()

	DefaultEngine = NewEngine(
		config.GetString("id"),
		config.GetString("bind"),
		config.GetString("advertise_addr"),
		config.GetString("server_addr"),
		int32(config.GetInt("worker")))
}

const (
	Version = "v0.0.0 -- HEAD"
)

type Engine struct {
	sync.RWMutex

	stopCh chan error

	id            string
	advertiseAddr string
	serverAddr    string
	dispatcher    *dispatcher
	workerCount   int32

	serverClient pb.ServerClient

	gRPCServer *grpc.Server
}

func NewEngine(id, bind, advertiseAddr, serverAddr string, workerCount int32) *Engine {

	d := newDispatcher()
	e := &Engine{
		stopCh:      make(chan error),
		id:          id,
		serverAddr:  serverAddr,
		dispatcher:  d,
		workerCount: workerCount,
		gRPCServer:  grpc.NewServer(),
	}

	if err := e.initAdvertiseAddr(advertiseAddr); err != nil {
		panic(err) // FIXME: remove panic
	}

	pbWorker.RegisterWorkerServer(e.gRPCServer, e)

	return e
}

func (e *Engine) initAdvertiseAddr(advertiseAddr string) error {

	hostOrInterface, port, err := net.SplitHostPort(advertiseAddr)
	if err != nil {
		return err
	}

	ip := net.ParseIP(hostOrInterface)
	if ip == nil {
		if ip, err = loadInterfaceIPv4(hostOrInterface); err != nil {
			return err
		}
	}

	e.advertiseAddr = net.JoinHostPort(ip.String(), port)
	return nil
}

func Func(name string, fun Function) { DefaultEngine.Func(name, fun) }
func (e *Engine) Func(name string, fun Function) {
	e.dispatcher.addFunc(name, fun)
}

func Run() error { return DefaultEngine.Run() }
func (e *Engine) Run() error {

	log.Printf("async: Running %s - %s - %d", e.id, e.advertiseAddr, e.workerCount)
	e.dispatcher.PrintDebug()

	go func() {
		if err := e.connectWithRetry(); err != nil {
			e.stopCh <- err
			return
		}
	}()

	// TODO: start heartbeat

	// bind := config.GetString("bind")

	go func() {

		lis, err := net.Listen("tcp", e.advertiseAddr)
		if err != nil {
			e.stopCh <- err
			return
		}

		log.Printf("async: Serving gRPC API on %s", e.advertiseAddr)
		if err := e.gRPCServer.Serve(lis); err != nil {
			e.stopCh <- err
			return
		}
	}()

	err := <-e.stopCh
	return err
}

func (e *Engine) Exec(ctx context.Context, in *pbWorker.ExecRequest) (*pbWorker.ExecReply, error) {

	log.Printf("async: Exec: %s ", in.GetFunction())

	if err := e.dispatcher.dispatch(ctx, in.GetFunction(), nil, nil); err != nil {
		log.Printf("async: failed to dispatch %s: %v", in.GetFunction(), err)
		return nil, err
	}

	return &pbWorker.ExecReply{}, nil
}

func (e *Engine) Info(ctx context.Context, in *pbWorker.InfoRequest) (*pbWorker.InfoReply, error) {

	return &pbWorker.InfoReply{
		Id:           e.id,
		Version:      Version,
		MaxParallel:  e.workerCount,
		Capabilities: e.dispatcher.Capabilities(),
	}, nil
}

func (e *Engine) connectWithRetry() error {
	// TODO: backoff
	for i := 0; i <= 5; i++ {

		if err := e.connect(); err != nil {
			return err
		} else {
			// Connected !
			return nil
		}
	}

	return ErrCannotConnect
}

func (e *Engine) connect() error {

	log.Printf("async: connecting on %s", e.serverAddr)

	// Set up a connection to the server.
	conn, err := grpc.Dial(e.serverAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	client := pb.NewServerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.RegisterWorker(ctx, &pb.RegisterWorkerRequest{
		Address: e.advertiseAddr,
	})
	if err != nil {
		return err
	}

	e.Lock()
	defer e.Unlock()
	e.serverClient = client

	return nil
}
