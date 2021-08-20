package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	prefix := fmt.Sprintf("[TEST_APP: %v -> %v] ", os.Getppid(), os.Getpid())

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	count := 1

	fmt.Println(prefix, "awaiting signal")

	go func() {
		for {
			fmt.Println(prefix, "counter ", count)
			count++
			time.Sleep(time.Second)
		}
	}()

	<-done
	fmt.Println(prefix, "exiting")
}
