package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/qqzeng/tinydocker/cgroups"
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"github.com/qqzeng/tinydocker/container"
	"github.com/qqzeng/tinydocker/network"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volumeStr string,
	containerName string, imageName string, envSlice []string, nw string, portmapping []string) {
	id := randStringBytes(container.NameLength)
	if containerName == "" {
		containerName = id
	}
	parent, wp := container.NewParentProcess(tty, volumeStr, containerName, imageName, envSlice)
	if parent == nil {
		log.Error("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	/* record container information */
	cName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName, id, volumeStr)
	if err != nil {
		log.Errorf("Record container information error: %v", err)
	}

	cgroupManager := cgroups.NewCgroupManager("tinydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	/* setup network information */
	if nw != "" {
		network.Init()
		cInfo := &container.ContainerInfo{
			Id:          id,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portmapping,
		}
		if err := network.Connect(nw, cInfo); err != nil {
			log.Errorf("Fail to connect network : %v", err)
			return
		}
	}

	sendInitCommand(comArray, wp)

	if tty {
		parent.Wait()
		/* TODO: need to delete container information for detached container process. */
		deleteContainerInfo(cName)
		container.DeleteWorkSpace(volumeStr, containerName)
	} else {
		log.Infof("Pid of current running container is %v", parent.Process.Pid)
		//time.Sleep(3 * time.Second)
	}
}

func sendInitCommand(comArray []string, wp *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	wp.WriteString(command)
	wp.Close()
}

func randStringBytes(n int) string {
	letters := "0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func recordContainerInfo(containerPid int, comArray []string, containerName string,
	id string, volumeStr string) (string, error) {
	/* construct container struct. */
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(comArray, " ")
	containerInfo := &container.ContainerInfo{
		Pid:        strconv.Itoa(containerPid),
		Id:         id,
		Name:       containerName,
		Command:    command,
		CreateTime: createTime,
		Status:     container.RUNNING,
		Volume:     volumeStr,
	}
	containerBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container %v information error %v", containerName, err)
		return "", err
	}
	containerInfoStr := string(containerBytes)

	/* create saving directories. */
	containerSavedUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// TODO: why created already?
	//exists, _ := PathExists(containerSavedUrl)
	//if exists {
	//	return "", fmt.Errorf("container %s exists, please give another container name", containerName)
	//}
	if err := os.MkdirAll(containerSavedUrl, 0622); err != nil {
		return "", fmt.Errorf("create container saved directory failed, %v", err)
	}
	log.Infof("Create container saved directory %s", containerSavedUrl)
	containerSavedFile := containerSavedUrl + container.ConfigName
	savedFile, err := os.Create(containerSavedFile)
	if err != nil {
		return "", fmt.Errorf("create container saved file failed, %v", err)
	}
	defer savedFile.Close()

	/* write container information to file. */
	if _, err := savedFile.WriteString(containerInfoStr); err != nil {
		return "", fmt.Errorf("write container infomation to file failed, %v", err)
	}
	return containerName, nil
}

func deleteContainerInfo(containerName string) {
	containerSavedUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	exists, _ := PathExists(containerSavedUrl)
	if !exists {
		log.Errorf("Container %s not found, abort delete operation", containerName)
		return
	}
	if err := os.RemoveAll(containerSavedUrl); err != nil {
		log.Errorf("Remove container information error: %v", err)
	}
}

func PathExists(url string) (bool, error) {
	_, err := os.Stat(url)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
