package agent

import (
	"calculator/internal/task/taskpb"
	"calculator/pkg/loggers"
	"context"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc/backoff"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run() {
	logger := loggers.GetLogger("agent")
	n, _ := strconv.Atoi(getenv("COMPUTING_POWER", "10"))
	sem := make(chan struct{}, n)
	for {
		conn, err := grpc.NewClient(
			getenv("ORCH_ADDR", "localhost:50051"),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithConnectParams(
				grpc.ConnectParams{Backoff: backoff.DefaultConfig,
					MinConnectTimeout: 5 * time.Second},
			),
		)
		if err != nil {
			logger.Error("gRPC NewClient", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		client := taskpb.NewOrchestratorClient(conn)
		stream, err := client.GetTasks(context.Background(), &taskpb.Empty{})
		if err != nil {
			logger.Error("GetTasks", "err", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		for {
			task, err := stream.Recv()
			if err != nil {
				logger.Error("stream recv", "err", err)
				break
			}
			sem <- struct{}{}
			go func(t *taskpb.Task) {
				defer func() { <-sem }()
				time.Sleep(time.Duration(t.OperationTime) * time.Millisecond)
				res := calc(t.Arg1, t.Arg2, t.Operation)
				if _, err := client.SendResult(context.Background(), &taskpb.SolvedTask{
					Id:     t.Id,
					Result: res,
				}); err != nil {
					logger.Error("SendResult", "err", err)
				}
			}(task)
		}
		conn.Close()
		time.Sleep(5 * time.Second)
	}
}

func calc(a, b float64, op string) float64 {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		return a / b
	}
	return 0
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
