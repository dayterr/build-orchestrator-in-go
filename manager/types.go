package manager

import (
	"build-orchestrator-in-go/task"

	"github.com/golang-collections/collections/queue"
	"github.com/google/uuid"
)

type Manager struct {
	Pending       queue.Queue
	TaskDB        map[string][]*task.Task
	EventDB       map[string][]*task.TaskEvent
	Workers       []string
	WorkerTaskMap map[string]uuid.UUID
	TaskWorkerMap map[uuid.UUID]string
}
