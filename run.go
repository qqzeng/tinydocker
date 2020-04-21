package main

import (
	"encoding/json"
	"fmt"
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"github.com/qqzeng/tinydocker/container"
	"github.com/qqzeng/tinydocker/cgroups"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volumeStr string, containerName string) {
	parent, wp := container.NewParentProcess(tty, volumeStr, containerName)
	if parent == nil {
		log.Error("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	/* record container information */
	cName, err := recordContainerInfo(parent.Process.Pid, comArray, containerName)
	if err != nil {
		log.Errorf("Record container information error: %v", err)
	}

	sendInitCommand(comArray, wp)
	cgroupManager := cgroups.NewCgroupManager("tinydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	if tty {
		parent.Wait()
		/* TODO: need to delete container information for detached container process. */
		deleteContainerInfo(cName)
	} else {
		log.Infof("Pid of current running container is %v", parent.Process.Pid)
		//time.Sleep(3 * time.Second)
	}
	container.DeleteWorkSpace(RootUrl, MntUrl, volumeStr)
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

func recordContainerInfo(containerPid int, comArray []string, containerName string) (string, error) {
	/* construct container struct. */
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(comArray, " ")
	id := randStringBytes(container.NameLength)
	if containerName == "" {
		containerName = id
	}
	containerInfo := &container.ContainerInfo {
		Pid:		strconv.Itoa(containerPid),
		Id:			id,
		Name:		containerName,
		Command:	command,
		CreateTime:	createTime,
		Status:		container.RUNNING,
	}
	containerBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container %v information error %v", containerName, err)
		return "", err
	}
	containerInfoStr := string(containerBytes)

	/* create saving directories. */
	containerSavedUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	// TODO: why already created?
	//exists, _ := PathExists(containerSavedUrl)
	//if exists {
	//	return "", fmt.Errorf("container %s exists, please give another container name", containerName)
	//}
	if err := os.MkdirAll(containerSavedUrl, 0622); err != nil {
		return "", fmt.Errorf("create container saved directory failed, %v", err)
	}
	log.Infof("Create container saved directory %s", containerSavedUrl)
	containerSavedFile := containerSavedUrl + container.ConfigName
	savedFile, err := os.Create(containerSavedFile);
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

func ListContainers() {
	/* load container list information from specific directory. */
	containerSavedUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	containerSavedUrl = containerSavedUrl[:len(containerSavedUrl)-1]
	containerFiles, err := ioutil.ReadDir(containerSavedUrl)
	if err != nil {
		log.Errorf("Read container information directory error: %s", err)
		return
	}
	var containerInfoList []*container.ContainerInfo
	for _, cf := range containerFiles {
		tmpC, err := extractContainerInfo(cf)
		if err != nil {
			log.Errorf("Read container information error: %s", err)
			continue
		}
		containerInfoList = append(containerInfoList, tmpC)
	}

	/* output container information to stdout */
	wr := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprintf(wr, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATETIME\n")
	for _, item := range containerInfoList {
		fmt.Fprintf(wr, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreateTime,
		)
	}
	if err := wr.Flush(); err != nil {
		log.Errorf("Flush container information to stdout error : %v", err)
		return
	}
}

func extractContainerInfo(cf os.FileInfo) (*container.ContainerInfo, error) {
	containerLocation := fmt.Sprintf(container.DefaultInfoLocation, cf.Name())
	containerFile := containerLocation + container.ConfigName
	content, err := ioutil.ReadFile(containerFile)
	if err != nil {
		log.Errorf("Read container file %s error: %v", containerFile, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Unmarshal container json content error: %v", err)
		return nil, err
	}
	return &containerInfo, nil
}

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

