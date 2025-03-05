package global

import "sync"

var (
	TasksMap   sync.Map
	FuturesMap sync.Map
)
