package container

import (
	"os"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"syscall"
)

const (
	rootUrl = "/root/"
	mntUrl = "/root/mnt/"
)

func NewParentProcess(tty bool) (*exec.Cmd, *os.File){
	rp, wp, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = [] *os.File{rp}
	NewWorkSpace(rootUrl, mntUrl)
	cmd.Dir = mntUrl
	return cmd, wp
}

func NewPipe() (*os.File, *os.File, error) {
	if rp, wp, err := os.Pipe(); err != nil {
		return nil, nil, err
	} else {
		return rp, wp, nil
	}
}

func NewWorkSpace(rootUrl string, mntUrl string) {
	CreateReadOnlyLayer(rootUrl)
	CreateWriteLayer(rootUrl)
	CreateMountPoint(rootUrl)
}

func CreateReadOnlyLayer(rootUrl string) {
	busyboxUrl := rootUrl + "busybox/"
	busyboxTarUrl := rootUrl + "busybox.tar"
	exist, err := PathExists(busyboxUrl)
	if err != nil {
		log.Infof("fail to judge whether directory %v exists: %v", busyboxUrl, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			log.Errorf("fail to create directory %v", busyboxUrl)
			return
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			log.Errorf("fail to untar %v", busyboxTarUrl)
		}
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

func CreateWriteLayer(rootUrl string) {
	writeLayerUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeLayerUrl, 0777); err != nil {
		log.Errorf("fail to create directory %s : %v", writeLayerUrl, err)
	}
}

func CreateMountPoint(rootUrl string) {
	mntUrl := rootUrl + "mnt/"
	exist, err := PathExists(mntUrl)
	if err != nil {
		log.Errorf("Error %v", err)
		return
	}
	if exist == false {
		if err := os.Mkdir(mntUrl, 0777); err != nil {
			log.Errorf("fail to create directory %s : %v", mntUrl, err)
		}
	}
	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount readonlyLayer and writeLayer error: %v", err)
	}
}

func DeleteWorkSpace(rootUrl string, mntUrl string) {
	DeleteMountPoint(rootUrl, mntUrl)
	DeleteWriteLayer(rootUrl)
}

func DeleteMountPoint(rootUrl string, mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount point error: %v", err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("remove mount point %s error: %v", mntUrl, err)
	}
}

func DeleteWriteLayer(rootUrl string) {
	writeLayerUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeLayerUrl); err != nil {
		log.Errorf("remove writeLayer %s error: %v", writeLayerUrl, err)
	}
}

