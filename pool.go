package gopool

import (
	"fmt"
	"strconv"
	"time"
)

type Task interface{}

// TaskHandleFunc business handle function
type TaskHandleFunc func(Task interface{}, workerName string) error

// SuccessCallback task handle success callback
type SuccessCallback func(Task interface{}, startTime time.Time, endTime time.Time)

// ErrorCallback task handle failed callback
type ErrorCallback func(Task interface{}, startTime time.Time, endTime time.Time, err error)

// HandleFunctions handle function collection
type HandleFunctions struct {
	TaskHandleFunc  TaskHandleFunc
	SuccessCallback SuccessCallback
	ErrorCallback   ErrorCallback
}

// Worker actual task performer
type Worker struct {
	Id         int
	Name       string
	Func       HandleFunctions
	WorkerPool chan chan Task
	InTask     chan Task
	QuitSignal chan bool
}

// CreateWorker Create a worker and add it to the pool.
func CreateWorker(workerPool chan chan Task, handFunctions HandleFunctions, id int, name string) Worker {
	return Worker{
		Id:         id,
		Func:       handFunctions,
		WorkerPool: workerPool,
		Name:       name + "-" + strconv.Itoa(id),
		InTask:     make(chan Task),
		QuitSignal: make(chan bool),
	}
}

// Start worker
func (w Worker) Start() {
	go func() {
		var msg Task
		defer func() {
			if r := recover(); r != nil {
				w.Func.ErrorCallback(msg, time.Now(), time.Now(), fmt.Errorf("fatal error in task: %s", r))
			}
		}()
		for {
			select {
			case <-w.QuitSignal:
				break
			case w.WorkerPool <- w.InTask:
				msg = <-w.InTask
				timeTrackStart := time.Now()
				err := w.Func.TaskHandleFunc(msg, w.Name)
				timeTrackStop := time.Now()
				if err != nil && w.Func.ErrorCallback != nil {
					w.Func.ErrorCallback(msg, timeTrackStart, timeTrackStop, err)
				} else if err == nil && w.Func.SuccessCallback != nil {
					w.Func.SuccessCallback(msg, timeTrackStart, timeTrackStop)
				}
			}
		}
	}()
}

// Stop the worker.
func (w Worker) Stop() {
	w.QuitSignal <- true
}

// CoroutinePool Processor structure
type CoroutinePool struct {
	workerPool    chan chan Task
	workerList    []Worker
	DecoderParams HandleFunctions
	Name          string
}

// CreateCoroutinePool Create a task coroutine pool which is going to create all the workers and set-up the pool.
func CreateCoroutinePool(numWorkers int, handlerParams HandleFunctions, name string) CoroutinePool {
	pool := CoroutinePool{
		workerPool:    make(chan chan Task),
		workerList:    make([]Worker, numWorkers),
		DecoderParams: handlerParams,
		Name:          name,
	}
	for i := 0; i < numWorkers; i++ {
		worker := CreateWorker(pool.workerPool, handlerParams, i, name)
		pool.workerList[i] = worker
	}
	return pool
}
func (p CoroutinePool) Start() {
	for _, worker := range p.workerList {
		worker.Start()
	}
}
func (p CoroutinePool) Stop() {
	for _, worker := range p.workerList {
		worker.Stop()
	}
}
func (p CoroutinePool) PublishTask(task Task) {
	taskChannel := <-p.workerPool
	taskChannel <- task
}
