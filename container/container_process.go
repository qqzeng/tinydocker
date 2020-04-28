package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
)

var (
	RootUrl = "/root"
	MntUrl  = "/root/mnt/%s"
	WriteLayer = "/root/writeLayer/%s"
)

type ContainerInfo struct {
	Pid			string `json:"pid"` 			/* the init process pid in host machine. */
	Id			string `json:"id"` 				/* the id of container. */
	Name		string `json:"name"`			/* the name of container */
	Command		string `json:"command"`			/* the executing command of init process in container */
	CreateTime	string `json:"createTime"`		/* the create time of container */
	Status		string `json:"status"`			/* the status of container */
	Volume 		string `json:"volume"`			/* the mounted volume of container */
	PortMapping []string `json:"portmapping"`	/* the port mapping of container */
}

const (
	RUNNING  			string = "running"
	STOP  	 			string = "stopped"
	EXIT  	 			string = "exited"
	DefaultInfoLocation string = "/var/run/tinydocker/%s/"
	ConfigName			string = "config.json"
	NameLength			int    = 10
	LogName				string = "container.log"
)

func NewParentProcess(tty bool, volumeStr string, containerName string, imageName string,
	envSlice []string) (*exec.Cmd, *os.File) {
	rp, wp, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init", containerName)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		/* redirect ouput of init process to a temporary file. */
		err, clf := createContainerLogFile(containerName)
		if err != nil {
			log.Errorf("Create container log file error : %v", err)
			// ...
		}
		cmd.Stdout = clf
		cmd.Stderr = clf
	}

	cmd.ExtraFiles = []*os.File{rp}
	cmd.Env = append(os.Environ(), envSlice...)
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	NewWorkSpace(volumeStr, imageName, containerName)
	return cmd, wp
}

func NewPipe() (*os.File, *os.File, error) {
	if rp, wp, err := os.Pipe(); err != nil {
		return nil, nil, err
	} else {
		return rp, wp, nil
	}
}

func createContainerLogFile(containerName string) (error, *os.File) {
	containerLogDir := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.MkdirAll(containerLogDir, 0622); err != nil {
		return fmt.Errorf("create log directory for container %s error : %v", containerName, err), nil
	}
	containerLogFile := containerLogDir + LogName
	clf, err := os.Create(containerLogFile)
	if err != nil {
		return fmt.Errorf("create log file for container %s error : %v", containerName, err), nil
	}
	return nil, clf
}