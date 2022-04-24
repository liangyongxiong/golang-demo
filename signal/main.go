package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    ch := make(chan os.Signal, 1)
    defer close(ch)
    signal.Notify(ch, syscall.SIGUSR1, syscall.SIGINT)
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
