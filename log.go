package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"github.com/qqzeng/tinydocker/container"
	log "github.com/Sirupsen/logrus"
)

func LogContainer(containerName string) {
	containerLogDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	containerLogFileDir := containerLogDir + container.LogName
	clf, err := os.Open(containerLogFileDir)
	if err != nil {
		log.Errorf("Open log file for container %v error : %v", containerName, err)
		return
	}
	containerLogContent, err := ioutil.ReadAll(clf)
	if err != nil {
		log.Errorf("Read log content for container %v error : %v", containerName, err)
		return
	}
	fmt.Fprintf(os.Stdout, string(containerLogContent))
}
