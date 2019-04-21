package taskTimer

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestNewTaskManager(t *testing.T) {
	taskM := NewTaskManager(TaskManagerLoggerOption(func(name string, level LEVEL, i interface{}) {
		if level == WARN {
			fmt.Println(level, i, "my log")
		}
	}))
	taskM.AddTask("test", "*/3 * * * * *", func() {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("do something")
	}, TaskMTimeOutOption(time.Second*10))
	taskM.AddTask("test1", "*/1 * * * * *", func() {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("test1 do something")
	}, TaskMTimeOutOption(time.Second*10))
	taskM.Start()
	go func() {
		time.Sleep(5 * time.Second)
		taskM.Stop()
	}()
	taskM.Wait()
}

func TestStopOneTask(t *testing.T) {
	taskM := NewTaskManager(TaskManagerLoggerOption(func(name string, level LEVEL, i interface{}) {
		log.Println(level, i, "my log")
	}))
	taskM.AddTask("test", "*/3 * * * * *", func() {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("test do something")
	}, TaskMTimeOutOption(time.Second*10))
	taskM.AddTask("test1", "*/1 * * * * *", func() {
		time.Sleep(time.Millisecond * 500)
		fmt.Println("test1 do something")
	}, TaskMTimeOutOption(time.Second*10))

	go func() {
		time.Sleep(time.Second * 3)
		taskM.StopOneTask("test1")
	}()

	taskM.Start()
	go func() {
		time.Sleep(5 * time.Second)

	}()
	taskM.Wait()
}
