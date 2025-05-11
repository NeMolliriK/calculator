package rpc

import (
	taskpb "calculator/internal/task"
	"calculator/pkg/global"
	"calculator/pkg/loggers"
	"context"
	"errors"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
)

type server struct {
	taskpb.UnimplementedOrchestratorServer
}

func (s *server) GetTasks(_ *taskpb.Empty, stream taskpb.Orchestrator_GetTasksServer) error {
	for {
		var sent bool
		global.TasksMap.Range(func(key, value any) bool {
			task := value.(*global.Task)
			if err := stream.Send(&taskpb.Task{
				Id:            task.ID,
				Arg1:          task.Arg1,
				Arg2:          task.Arg2,
				Operation:     task.Operation,
				OperationTime: int32(task.OperationTime),
			}); err != nil {
				return false
			}
			global.TasksMap.Delete(key)
			sent = true
			return false // однократно за итерацию цикла
		})
		if !sent {
			// если задач нет — небольшая пауза
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (s *server) SendResult(ctx context.Context, in *taskpb.SolvedTask) (*taskpb.Empty, error) {
	if f, ok := global.FuturesMap.Load(in.GetId()); ok {
		f.(*global.Future).SetResult(in.GetResult())
	}
	return &taskpb.Empty{}, nil
}

func Run(ctx context.Context) (func(context.Context) error, error) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	taskpb.RegisterOrchestratorServer(grpcServer, &server{})
	go func() {
		if err := grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			loggers.GetLogger("orchestrator").Error("gRPC Serve", "err", err)
		}
	}()
	log.Println("gRPC server listening on :50051")
	return func(_ context.Context) error {
		grpcServer.GracefulStop()
		return nil
	}, nil
}
