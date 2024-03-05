package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	pb_svc_fixer "github.com/aglide100/ai-test/pb/svc/fixer"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"go.uber.org/zap"

	"github.com/aglide100/ai-test/pkg/cache"
	"github.com/aglide100/ai-test/pkg/controller"
	"github.com/aglide100/ai-test/pkg/gen"
	"github.com/aglide100/ai-test/pkg/logger"
	"github.com/aglide100/ai-test/pkg/model"
	"github.com/aglide100/ai-test/pkg/queue"
	fixer_server "github.com/aglide100/ai-test/pkg/svc/fixer/server"
	"github.com/rs/cors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	wsAddr         = flag.String("wsAddr", "0.0.0.0:9090", "websocket addrss")
	fixerAddr      = flag.String("fixerAddr", "0.0.0.0:50013", "grpc address")
	timeOutTime    = flag.Int("timeout", 60, "seconds")
	serverCrt      = flag.String("ccrtPath", "./keys/localhost.crt", "crt file location")
	serverKey      = flag.String("keyPath", "./keys/localhost.key", "ket file location")
	tls            = flag.Bool("isTls", false, "use tls")
	// serverCrt = flag.String("crtPath", "/run/secrets/crt-file", "docker secret crt file location")
	// serverKey = flag.String("keyPath", "/run/secrets/key-file", "docker secret ket file location")
)

func main() {
	if err := realMain(); err != nil {
		log.Printf("err :%v", err)
		os.Exit(1)
	}
}

func realMain() error {
	flag.Parse()

	logger.Info("timeout", zap.Any("timeout", timeOutTime))
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(*fixerAddr))
	if err != nil {
		return err
	}
	defer grpcListener.Close()

	var wait sync.WaitGroup
	wait.Add(1)

	doneJob := make(chan string, 10)
	readableRequest := make(chan *model.RequestData, 10)
	mutex := &sync.Mutex{}

	clientCache := cache.NewClientCache(time.Duration(*timeOutTime)*time.Second, mutex)
	// blobCache := cache.NewCache(time.Duration(5)*time.Minute, mutex, true)
	blobCache := cache.NewBlobCache(time.Duration(5)*time.Minute, mutex)
	resultCache := cache.NewCache(time.Duration(5)*time.Minute, mutex, true)
	waitingChannels := cache.NewCache(time.Duration(1)*time.Minute, mutex, true)

	taskAllocator := queue.NewTaskAllocator(mutex)

	token := os.Getenv("TOKEN")
	if len(token) == 0 {
		token = gen.RandStringRunes(20)
		logger.Info("token is nil, use random string", zap.String("token", token))
	} else {
		log.Printf("Token : %s", token)
	}

	fixerSrv := fixer_server.NewFixerServiceServer(taskAllocator, token, doneJob, readableRequest, resultCache, blobCache, *timeOutTime)

	size := 1024 * 1024 * 50
	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(size),
		grpc.MaxRecvMsgSize(size))

	pb_svc_fixer.RegisterFixerServiceServer(grpcServer, fixerSrv)

	wg, _ := errgroup.WithContext(context.Background())

	wg.Go(func() error {
		log.Printf("Starting grpcServer at: %s", *fixerAddr)
		err := grpcServer.Serve(grpcListener)
		if err != nil {
			log.Fatalf("failed to serve: %v", err)
			return err
		}

		return nil
	})

	wg.Go(func() error {
		duration, _ := time.ParseDuration("10s")
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for range ticker.C {
			taskAllocator.CleanUp(*timeOutTime)
		}

		return nil
	})

	grpcDialOption := grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(size), grpc.MaxCallSendMsgSize(size))

	conn, err := grpc.DialContext(
		context.Background(),
		*fixerAddr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcDialOption,
	)

	if err != nil {
		return err
	}

	gwmux := runtime.NewServeMux()
	err = pb_svc_fixer.RegisterFixerServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		return err
	}

	ctl := controller.NewWsController(token, taskAllocator, 2, mutex, doneJob, clientCache, resultCache, waitingChannels, readableRequest)

	withCors := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"ACCEPT", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
	}).Handler(gwmux)

	http.HandleFunc("/connect/", addHeaders(http.HandlerFunc(ctl.CreateConnection)))
	http.HandleFunc("/", withCors.ServeHTTP)

	wg.Go(func() error {
		log.Printf("Starting http srv at: %s", *wsAddr)

		if *tls {
			logger.Info("serve by tls")
			err := http.ListenAndServeTLS(*wsAddr, *serverCrt, *serverKey, nil)
			if err != nil {
				return err
			}
		} else {
			err := http.ListenAndServe(*wsAddr, nil)
			if err != nil {
				return err
			}
		}

		return nil
	})

	wg.Go(func() error {
		ctl.Watcher()
		return nil
	})

	return wg.Wait()
}

func addHeaders(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		h.ServeHTTP(w, r)
	}
}
