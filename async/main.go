package main

import (
	"fmt"
	machinery "github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"github.com/RichardKnop/machinery/v1/tasks"
	"os"
	"os/signal"
	"syscall"
	"time"

	"async/call"
)

var cnf = &config.Config{
	Broker:        "redis://127.0.0.1:6379/0",
	DefaultQueue:  "default_queue",
	ResultBackend: "eager",
	//ResultBackend: "redis://127.0.0.1:6379/0",
}

func InitServer() *machinery.Server {
	server, err := machinery.NewServer(cnf)
	if err != nil {
		return nil
	}
	return server
}

func LaunchWorker(server *machinery.Server) {
	fmt.Println("worker initing")
	worker := server.NewWorker("worker_name", 10)
	fmt.Println("worker inited")
	err := worker.Launch()
	fmt.Println("worker launched")
	if err != nil {
		fmt.Println(err)
	}
}

func SendTask(server *machinery.Server) {
	server.RegisterTask("add", call.Add)
	server.RegisterTask("multiply", call.Multiply)

	signature := &tasks.Signature{
		Name: "add",
		Args: []tasks.Arg{
			{
				Type:  "int64",
				Value: 1,
			},
			{
				Type:  "int64",
				Value: 1,
			},
		},
	}

	asyncResult, err := server.SendTask(signature)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("id=%s, state=%s\n", asyncResult.GetState().TaskUUID, asyncResult.GetState().State)
	results, _ := asyncResult.Get(time.Duration(time.Millisecond * 5))
	fmt.Printf("id=%s, state=%s\n", asyncResult.GetState().TaskUUID, asyncResult.GetState().State)
	for _, result := range results {
		fmt.Printf("value=%v\n", result.Interface())
	}

}

func main() {
	server := InitServer()
	go LaunchWorker(server)
	SendTask(server)

	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, syscall.SIGINT)
	go func() {
		for {
			sig := <-ch
			switch sig {
			case syscall.SIGINT:
				fmt.Printf("\ncatch interrupt signal: %v\n", sig)
				time.Sleep(5 * time.Second)
				os.Exit(0)
			default:
				fmt.Printf("catch other signal: %v\n", sig)
			}
		}
	}()

    // wait forever
    select {}
}
