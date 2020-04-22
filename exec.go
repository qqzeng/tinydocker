package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"github.com/qqzeng/tinydocker/container"
	log "github.com/Sirupsen/logrus"
	_ "github.com/qqzeng/tinydocker/nsenter"
)
/* TODO: why are there two `sh` process? */
func ExecContainer(containerName string, comArray []string) {
	cPid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Error("Finding pid of container process error: %v", err)
		return
	}
	containerInfo, err := getContainerByName(containerName)
	if err != nil {
		log.Errorf("Get container name %s error : %v", containerName, err)
		return
	}
	if containerInfo.Status != container.RUNNING {
		log.Errorf("Can only exec running container name")
		return
	}
	log.Infof("The pid of container process is %s", cPid)
	if comArray == nil || len(comArray) == 0 {
		log.Error("Invalid command array for executing : %v", err)
		return
	}
	comStr := strings.Join(comArray, " ")
	log.Infof("The executing command of container process is %s", comStr)

	command := exec.Command("/proc/self/exe", "exec")
	/* NOTE: this sounds like making no sense.*/
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err1 := os.Setenv(ENV_EXEC_PID, cPid)
	err2 := os.Setenv(ENV_EXEC_COMMAND, comStr)
	if err1 != nil || err2 != nil {
		log.Errorf("Setting environment command error : %v", err)
	}

	containerEnvs := getEnvsByPid(cPid)
	command.Env = append(os.Environ(), containerEnvs...)

	if err = command.Run(); err != nil {
		log.Errorf("Run command error : %v", err)
	}
}

func getContainerPidByName(containerName string) (string, error) {
	if containerName == "" {
		return "", fmt.Errorf("invalid container name %s", containerName)
	}
	containerSavedDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	containerInfoFileDir := containerSavedDir + container.ConfigName
	clf, err := os.Open(containerInfoFileDir)
	if err != nil {
		return "", fmt.Errorf("open container saved file for container %s error : %v", containerName, err)
	}
	containerContent, err := ioutil.ReadAll(clf)
	if err != nil {
		return "", fmt.Errorf("read container saved file for container %s error : %v", containerName, err)
	}
	var containerInfo container.ContainerInfo
	err = json.Unmarshal(containerContent, &containerInfo)
	if err != nil {
		return "", fmt.Errorf("unmarshal container content for container %s error : %v", containerName, err)
	}
	return string(containerInfo.Pid), nil
}

func getEnvsByPid(pid string) []string {
	envPath := fmt.Sprintf("/proc/%s/environ", pid)
	envBytes, err := ioutil.ReadFile(envPath)
	if err != nil {
		log.Errorf("Read environment variables file %s error : %v", envPath, err)
		return nil
	}
	envs := strings.Split(string(envBytes), "\u0000")
	return envs
}