package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/qqzeng/tinydocker/container"
	"os"
)

/*  TODO: clear the volume directory of container. */
func RemoveContainer(containerName string) {
	containerInfo, err := getContainerByName(containerName)
	if err != nil {
		log.Errorf("Get container name %s error : %v", containerName, err)
		return
	}
	if containerInfo.Status != container.STOP && containerInfo.Status != container.EXIT {
		log.Errorf("Can not remove %s container %s", containerInfo.Status, containerName)
		return
	}
	containerSavedDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(containerSavedDir); err != nil {
		log.Errorf("Remove container name %s error : %v", containerName, err)
		return
	}
}
