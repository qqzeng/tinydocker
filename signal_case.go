package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	hokOfProcessExit()
	binary, lookErr := exec.LookPath("ls")
	if lookErr != nil {
		panic(lookErr)
	}
	var cmdArray = [2]string {"ls", "-l"}
	if err := syscall.Exec(binary, cmdArray[:], os.Environ()); err != nil {
		fmt.Printf("error : %v\n", err.Error())
	}
	//if _, err := exec.Command("top", "-b").CombinedOutput(); err != nil {
	//	fmt.Printf("error : %v\n", err.Error())
	//}
}

func hokOfProcessExit() {
	var stopLock sync.Mutex
	stop := false
	signalChan := make(chan os.Signal, 1)
	go func() {
		fmt.Println("Waiting for container process exit..")
		<-signalChan
		stopLock.Lock()
		stop = true
		stopLock.Unlock()
		fmt.Println("Cleaning before stop...")
		os.Exit(0)
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
}