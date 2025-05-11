package rpc

import (
	"context"
	"testing"
	"time"

	"calculator/internal/global"
	"calculator/internal/task/taskpb"

	"google.golang.org/grpc/metadata"
)

type fakeStream struct {
	Sent []*taskpb.Task
	ctx  context.Context
}

func (f *fakeStream) Send(task *taskpb.Task) error {
	f.Sent = append(f.Sent, task)
	return nil
}
func (f *fakeStream) Context() context.Context        { return f.ctx }
func (f *fakeStream) SetHeader(md metadata.MD) error  { return nil }
func (f *fakeStream) SendHeader(md metadata.MD) error { return nil }
func (f *fakeStream) SetTrailer(md metadata.MD)       {}
func (f *fakeStream) SendMsg(m interface{}) error     { return nil }
func (f *fakeStream) RecvMsg(m interface{}) error     { return nil }

func clearMaps() {
	global.TasksMap.Range(func(key, _ any) bool {
		global.TasksMap.Delete(key)
		return true
	})
	global.FuturesMap.Range(func(key, _ any) bool {
		global.FuturesMap.Delete(key)
		return true
	})
}

func TestSendResult_SetsFutureResult(t *testing.T) {
	clearMaps()
	fut := global.NewFuture()
	global.FuturesMap.Store("task1", fut)

	srv := &server{}
	_, err := srv.SendResult(context.Background(), &taskpb.SolvedTask{Id: "task1", Result: 3.14})
	if err != nil {
		t.Fatalf("SendResult returned error: %v", err)
	}

	res := fut.Get()
	if res != 3.14 {
		t.Errorf("Future.Get() = %v, want 3.14", res)
	}
}

func TestGetTasks_NoTasks_ShutdownImmediately(t *testing.T) {
	clearMaps()
	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	shutdownCancel()

	srv := &server{shutdownCtx: shutdownCtx}
	stream := &fakeStream{ctx: context.Background()}

	err := srv.GetTasks(&taskpb.Empty{}, stream)
	if err != nil {
		t.Fatalf("GetTasks returned error: %v", err)
	}
	if len(stream.Sent) != 0 {
		t.Errorf("Sent = %d tasks, want 0", len(stream.Sent))
	}
}

func TestGetTasks_SendsOneTaskThenStops(t *testing.T) {
	clearMaps()
	task := &global.Task{ID: "t1", Arg1: 1, Arg2: 2, Operation: "+", OperationTime: 0}
	global.TasksMap.Store(task.ID, task)

	shutdownCtx, shutdownCancel := context.WithCancel(context.Background())
	defer shutdownCancel()
	go func() {
		time.Sleep(50 * time.Millisecond)
		shutdownCancel()
	}()

	srv := &server{shutdownCtx: shutdownCtx}
	stream := &fakeStream{ctx: context.Background()}

	err := srv.GetTasks(&taskpb.Empty{}, stream)
	if err != nil {
		t.Fatalf("GetTasks returned error: %v", err)
	}
	if len(stream.Sent) != 1 {
		t.Fatalf("Sent = %d tasks, want 1", len(stream.Sent))
	}
	sent := stream.Sent[0]
	if sent.Id != task.ID || sent.Arg1 != task.Arg1 || sent.Arg2 != task.Arg2 || sent.Operation != task.Operation {
		t.Errorf("Sent task = %+v, want %+v", sent, task)
	}
}
