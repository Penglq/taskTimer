package taskTimer

import (
	"fmt"
	"time"
)

type TaskInterface interface {
	Run()
}

func NewTask(name, rule string, taskFunc func(), options ...TaskOption) *Task {
	t := &Task{
		Name:        name,
		TaskRule:    rule,
		Async:       false,
		TaskFunc:    taskFunc,
		StopCH:      make(chan struct{}, 1),
		TaskLog: func(name string, level LEVEL, s interface{}) {
			fmt.Println(name, level, s)
		},
	}
	for _, option := range options {
		option(t)
	}
	return t
}

type Task struct {
	Name        string
	Timer       *time.Timer
	TaskRule    string // * * * * * * 秒 分 时 日 月 星期
	Async       bool   // 是否异步执行;如果false,则在一次任务完成后才进行下一次任务记时
	TaskFunc    func()
	TaskLog     func(string, LEVEL, interface{})
	StopCH      chan struct{}
}

type TaskOption func(*Task)

func TaskLoggerOption(logger func(string, LEVEL, interface{})) TaskOption {
	return func(t *Task) {
		t.TaskLog = logger
	}
}

func TaskAsyncOption(async bool) TaskOption {
	return func(t *Task) {
		t.Async = async
	}
}

func (t *Task) SetTaskDesc(desc string) {
	t.TaskRule = desc
}

func (t *Task) Run() {
	t.Timer = time.NewTimer(time.Until(t.NextRunTime()))
	defer func() {
		t.Timer.Stop()
	}()
	for {
		select {
		case <-t.Timer.C:
			t.Timer.Reset(time.Until(t.NextRunTime()))
			t.do()
		case <-t.StopCH:
			t.StopCH <- struct{}{} // 给外部发信号
			return
		}
	}
}

func (t *Task) Stop() {
	t.StopCH <- struct{}{}
}

func (t *Task) CheckStop() (chan struct{}) {
	return t.StopCH
}

func (t *Task) NextRunTime() time.Time {
	sc, err := Parse(t.TaskRule)
	if err != nil {
		t.TaskLog(t.Name, WARN, "解析时间规则错误")
		return time.Unix(time.Now().Unix()-1, 0) // 获取过去时间
	}
	return sc.Next(time.Now())
}

func (t *Task) SetLogger(logger func(string, LEVEL, interface{})) {
	t.TaskLog = logger
}

func (t *Task) do() {
	if t.Async {
		go t.taskMiddleware(t.TaskFunc)()
	} else {
		t.taskMiddleware(t.TaskFunc)()
	}
}

func (t *Task) recover() {
	defer func() {
		if e := recover(); e != nil {
			t.TaskLog(t.Name, WARN, fmt.Sprintf("%v", e))
		}
	}()
}

func (t *Task) taskMiddleware(f func()) func() {
	return func() {
		defer func(TN time.Time) {
			t.TaskLog(t.Name, INFO, time.Since(TN))
		}(time.Now())
		f()
	}
}
