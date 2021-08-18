package gopool

import (
	"fmt"
	"testing"
	"time"
)

func TestCreateProcessor(t *testing.T) {
	//开启协程池
	processor := CreateCoroutinePool(3, HandleFunctions{
		TaskHandleFunc: func(Task interface{}, processorName string) error {
			return nil
		},
		SuccessCallback: func(Task interface{}, startTime time.Time, endTime time.Time) {

		},
		ErrorCallback: func(Task interface{}, startTime time.Time, endTime time.Time, err error) {

		},
	}, "externalTaskRoutine")
	processor.Start()
	processor.PublishTask(nil)
	processor.PublishTask(nil)
	processor.PublishTask(nil)
	processor.PublishTask(nil)
	fmt.Print(len(processor.workerPool))
}
