package container

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("run container get user command error, cmdArray is nil")
	}
	setupMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		logrus.Errorf(err.Error())
	}
	return nil
}

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func pivotRoot2(rootfs string) error {
	// While the documentation may claim otherwise, pivot_root(".", ".") is
	// actually valid. What this results in is / being the new root but
	// /proc/self/cwd being the old root. Since we can play around with the cwd
	// with pivot_root this allows us to pivot without creating directories in
	// the rootfs. Shout-outs to the LXC developers for giving us this idea.

	oldroot, err := syscall.Open("/", syscall.O_DIRECTORY|syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(oldroot)

	newroot, err := syscall.Open(rootfs, syscall.O_DIRECTORY|syscall.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(newroot)

	// Change to the new root so that the pivot_root actually acts on it.
	if err := syscall.Fchdir(newroot); err != nil {
		return err
	}

	if err := syscall.PivotRoot(".", "."); err != nil {
		return fmt.Errorf("pivot_root %s", err)
	}

	// Currently our "." is oldroot (according to the current kernel code).
	// However, purely for safety, we will fchdir(oldroot) since there isn't
	// really any guarantee from the kernel what /proc/self/cwd will be after a
	// pivot_root(2).

	if err := syscall.Fchdir(oldroot); err != nil {
		return err
	}

	// Make oldroot rslave to make sure our unmounts don't propagate to the
	// host (and thus bork the machine). We don't use rprivate because this is
	// known to cause issues due to races where we still have a reference to a
	// mount while a process in the host namespace are trying to operate on
	// something they think has no mounts (devicemapper in particular).
	if err := syscall.Mount("", ".", "", syscall.MS_SLAVE|syscall.MS_REC, ""); err != nil {
		return err
	}
	// Preform the unmount. MNT_DETACH allows us to unmount /proc/self/cwd.
	if err := syscall.Unmount(".", syscall.MNT_DETACH); err != nil {
		return err
	}

	// Switch back to our shiny new root.
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %s", err)
	}
	return nil
}

/**
  refer: http://man7.org/linux/man-pages/man2/pivot_root.2.html
 */
func pivotRoot(root string) error {

	/* Ensure that 'new_root' is a mount point */
	if err := syscall.Mount(root, root, "bind", uintptr(syscall.MS_BIND | syscall.MS_REC), ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error : %v", err)
	}
	/* Create directory to which old root will be pivoted */
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return fmt.Errorf("make pivotDir error : %v", err)
	}
	/* pivot the root filesystem */
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	/* Switch the current working directory to "/" */
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("change director error : %v", err)
	}
	/* Unmount old root and remove mount point */
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivotDir error : %v", err)
	}
	return os.Remove(pivotDir)
}

func setupMount()  {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current working directory error: %v", err)
		return
	}
	log.Infof("Current working directory is %v", pwd)
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		log.Errorf("mount / error: %v", err)
		return
	}
	pivotRoot(pwd)
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(syscall.MS_NOSUID | syscall.MS_STRICTATIME), "mode=755")
}