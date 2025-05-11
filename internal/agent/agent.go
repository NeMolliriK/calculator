package agent

import (
	taskpb "calculator/internal/task"
	"calculator/pkg/loggers"
	"context"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

func Run() {
	logger := loggers.GetLogger("general")
	ctx := context.Background()

	conn, err := grpc.DialContext(
		ctx,
		getenv("ORCH_ADDR", "localhost:50051"),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		logger.Error("gRPC dial", "err", err)
		return
	}
	defer conn.Close()
	client := taskpb.NewOrchestratorClient(conn)

	stream, err := client.GetTasks(ctx, &taskpb.Empty{})
	if err != nil {
		logger.Error("GetTasks", "err", err)
		return
	}

	n, _ := strconv.Atoi(getenv("COMPUTING_POWER", "10"))
	sem := make(chan struct{}, n)

	for {
		task, err := stream.Recv()
		if err != nil {
			logger.Error("stream recv", "err", err)
			return
		}
		sem <- struct{}{}
		go func(t *taskpb.Task) {
			defer func() { <-sem }()
			time.Sleep(time.Duration(t.OperationTime) * time.Millisecond)
			res := calc(t.Arg1, t.Arg2, t.Operation)
			if _, err := client.SendResult(ctx, &taskpb.SolvedTask{Id: t.Id, Result: res}); err != nil {
				logger.Error("SendResult", "err", err)
			}
		}(task)
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
