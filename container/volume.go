package container

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(rootUrl string, mntUrl string, volumeStr string) {
	CreateReadOnlyLayer(rootUrl)
	CreateWriteLayer(rootUrl)
	CreateMountPoint(rootUrl)
	valid, volumeUrls := ExtractVolumeParameter(volumeStr)
	if valid {
		MountVolume(rootUrl, mntUrl, volumeUrls)
	}
}

func MountVolume(rootUrl string, mntUrl string, volumeUrls []string) {
	hostUrl := volumeUrls[0]
	exist, _ := PathExists(hostUrl)
	if !exist {
		if err := os.Mkdir(hostUrl, 0777); err != nil {
			log.Infof("Mkdir host volume url %s error : %v", hostUrl, err)
		}
	}
	containerUrl := mntUrl + volumeUrls[1]
	if err := os.Mkdir(containerUrl, 0777); err != nil {
		log.Infof("Mkdir container volume url %s error : %v", containerUrl, err)
	}
	dirs := "dirs=" + hostUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume error : %v", err)
	}
}

func ExtractVolumeParameter(volumeStr string) (bool, []string) {
	var volumeUrls []string
	volumeUrls = strings.Split(volumeStr, ":")
	if volumeUrls != nil && len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
		return true, volumeUrls
	}
	return false, volumeUrls
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

func DeleteWorkSpace(rootUrl string, mntUrl string, volumeStr string) {
	valid, volumeUrls := ExtractVolumeParameter(volumeStr)
	if valid {
		DeleteMountPointWithVolume(rootUrl, mntUrl, volumeUrls)
	} else {
		DeleteMountPoint(rootUrl, mntUrl)
	}
	DeleteWriteLayer(rootUrl)
}

func DeleteMountPointWithVolume(rootUrl string, mntUrl string, volumeUrls []string) {
	/* unmount container volume. */
	containerUrl := mntUrl + volumeUrls[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Unmount volume %s error : %v", containerUrl, err)
	}
	/* unmount total mount point volume. */
	cmd = exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Unmount volume %s error : %v", mntUrl, err)
	}
	/* remove total mount point directory. */
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Remove volume %s error : %v", mntUrl, err)
	}
}

func DeleteMountPoint(rootUrl string, mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume %s error: %v", mntUrl, err)
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
