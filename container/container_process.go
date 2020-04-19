package container

import (
	"os"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"syscall"
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
	return cmd, wp
}

func NewPipe() (*os.File, *os.File, error) {
	if rp, wp, err := os.Pipe(); err != nil {
		return nil, nil, err
	} else {
		return rp, wp, nil
	}
}

//func NewParentProcess(tty bool, command string) *exec.Cmd{
//
//	args := []string{"init", command}
//	cmd := exec.Command("/proc/self/exe", args...)
//	cmd.SysProcAttr = &syscall.SysProcAttr{
//		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
//	}
//	if tty {
//		cmd.Stdin = os.Stdin
//		cmd.Stdout = os.Stdout
//		cmd.Stderr = os.Stderr
//	}
//	return cmd
//}