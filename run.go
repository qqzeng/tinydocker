package main

import (
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"github.com/qqzeng/tinydocker/container"
	"github.com/qqzeng/tinydocker/cgroups"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"
)

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, volumeStr string) {
	parent, wp := container.NewParentProcess(tty, volumeStr)
	if parent == nil {
		log.Error("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	sendInitCommand(comArray, wp)
	cgroupManager := cgroups.NewCgroupManager("tinydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	if tty {
		parent.Wait()
		container.DeleteWorkSpace(RootUrl, MntUrl, volumeStr)
	}
	os.Exit(-1)
}

func sendInitCommand(comArray []string, wp *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	wp.WriteString(command)
	wp.Close()
}