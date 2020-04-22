package main

import (
	"encoding/json"
	"fmt"
	"github.com/qqzeng/tinydocker/container"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

func StopContainer(containerName string) {
	cPid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Invalid container name %s : %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(cPid)
	if err != nil {
		log.Errorf("Invalid container pid %s : %v", cPid, err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error : %v", cPid, err)
		return
	}
	containerInfo, err := getContainerByName(containerName)
	if err != nil {
		log.Errorf("Get container name %s error : %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = ""
	updatedContainerBytes, err := json.Marshal(containerInfo);
	if err != nil {
		log.Errorf("Remarshal container name %s error : %v", containerName, err)
		return
	}
	containerSavedDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	containerInfoFileDir := containerSavedDir + container.ConfigName
	if err := ioutil.WriteFile(containerInfoFileDir, updatedContainerBytes, 0622); err != nil {
		log.Errorf("Write updated container content name for %s error : %v", containerName, err)
		return
	}
}

func getContainerByName(containerName string) (*container.ContainerInfo, error) {
	if containerName == "" {
		return nil, fmt.Errorf("invalid container name %s", containerName)
	}
	containerSavedDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	containerInfoFileDir := containerSavedDir + container.ConfigName
	clf, err := os.Open(containerInfoFileDir)
	if err != nil {
		return nil, fmt.Errorf("open container saved file for container %s error : %v", containerName, err)
	}
	containerContent, err := ioutil.ReadAll(clf)
	if err != nil {
		return nil, fmt.Errorf("read container saved file for container %s error : %v", containerName, err)
	}
	var containerInfo container.ContainerInfo
	err = json.Unmarshal(containerContent, &containerInfo)
	if err != nil {
		return nil, fmt.Errorf("unmarshal container content for container %s error : %v", containerName, err)
	}
	return &containerInfo, nil
}
