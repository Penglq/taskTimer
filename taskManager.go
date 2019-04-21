package taskTimer

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LEVEL string

const (
	Manager = "manager"
	INFO    = "info"
	WARN    = "warn"
)

type TaskManagerInterface interface {
	AddTask(string, func())
	Start()
}

func NewTaskManager(options ...TaskManagerOption) *TaskManager {
	m := &TaskManager{
		Tasks:  make(map[string]*TaskAttr),
		WG:     new(sync.WaitGroup),
		StopCH: make(chan struct{}, 1),
		Logger: func(name string, level LEVEL, s interface{}) {
			fmt.Println(name, level, s)
		},
	}
	for _, option := range options {
		option(m)
	}
	return m
}

type TaskManager struct {
	Tasks  map[string]*TaskAttr
	WG     *sync.WaitGroup
	Ctx    context.Context
	Cancel context.CancelFunc
	StopCH chan struct{}
	Logger func(string, LEVEL, interface{})
}

type TaskManagerOption func(*TaskManager)

func TaskManagerLoggerOption(logger func(string, LEVEL, interface{})) TaskManagerOption {
	return func(t *TaskManager) {
		t.Logger = logger
	}
}

type TaskAttr struct {
	Task    *Task
	TimeOut time.Duration
}
type TaskAttrOption func(*TaskAttr)

func TaskMTimeOutOption(timeOut time.Duration) TaskAttrOption {
	return func(t *TaskAttr) {
		t.TimeOut = timeOut
	}
}
func (t *TaskManager) SetLogger(logger func(string, LEVEL, interface{})) {
	t.Logger = logger
}

func (t *TaskManager) AddTask(name, rule string, f func(), options ...TaskAttrOption) {
	t.Tasks[name] = &TaskAttr{Task: NewTask(name, rule, f, TaskLoggerOption(t.Logger))}
	for _, task := range options {
		task(t.Tasks[name])
	}
}

func (t *TaskManager) Start() {
	t.WG.Add(len(t.Tasks))
	for _, task := range t.Tasks {
		t.Logger(Manager, INFO, task.Task.Name+` start`)
		go task.Task.Run()
	}
}

func (t *TaskManager) Stop() {
	for taskName, _ := range t.Tasks {
		t.Logger(Manager, INFO, taskName+` stop`)
		go t.Stopfunc(taskName)
	}
}

func (t *TaskManager) StopOneTask(taskName string) {
	t.Logger(Manager, INFO, taskName+` stop`)
	go t.Stopfunc(taskName)
}

func (t *TaskManager) Wait() {
	t.WG.Wait()
}

func (t *TaskManager) Stopfunc(taskName string) {
	timeOut := time.NewTimer(t.Tasks[taskName].TimeOut)
	defer func() {
		if !timeOut.Stop() {
			<-timeOut.C
		}
	}()
	t.Tasks[taskName].Task.Stop()
	select {
	case <-t.Tasks[taskName].Task.CheckStop():
	case <-timeOut.C:
		timeOut.Stop()
		t.Logger("", "", "time out")
	}
	t.WG.Done()
}
