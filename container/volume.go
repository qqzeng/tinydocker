package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(volumeStr string, imageName string, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	valid, volumeUrls := ExtractVolumeParameter(volumeStr)
	if valid {
		MountVolume(volumeUrls, containerName)
	}
}

func MountVolume(volumeUrls []string, containerName string) {
	hostUrl := volumeUrls[0]
	exist, _ := PathExists(hostUrl)
	if !exist {
		if err := os.Mkdir(hostUrl, 0777); err != nil {
			log.Errorf("Mkdir host volume url %s error : %v", hostUrl, err)
		}
	}
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntUrl + "/" +volumeUrls[1]
	if err := os.Mkdir(containerUrl, 0777); err != nil {
		log.Errorf("Mkdir container volume url %s error : %v", containerUrl, err)
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

func CreateReadOnlyLayer(imageName string) {
	imageUrl := RootUrl + "/" +  imageName + "/"
	imageTarUrl := RootUrl + "/" +  imageName + ".tar"
	exist, err := PathExists(imageUrl)
	if err != nil {
		log.Infof("Fail to judge whether directory %v exists: %v", imageUrl, err)
	}
	if exist == false {
		if err := os.MkdirAll(imageUrl, 0622); err != nil {
			log.Errorf("Create directory %s error : %v", imageUrl, err)
			return
		}
		if _, err := exec.Command("tar", "-xvf", imageTarUrl, "-C", imageUrl).CombinedOutput(); err != nil {
			log.Errorf("Fail to untar %v", imageTarUrl)
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

func CreateWriteLayer(containerName string) {
	writeLayerUrl := fmt.Sprintf(WriteLayer, containerName)
	exist, _ := PathExists(writeLayerUrl)
	if exist == true {
		if err := os.RemoveAll(writeLayerUrl); err != nil {
			log.Errorf("Remove exists writeLayer %s error : %v", writeLayerUrl, err)
			return
		}
	}
	if err := os.MkdirAll(writeLayerUrl, 0777); err != nil {
		log.Errorf("Fail to create directory %s : %v", writeLayerUrl, err)
	}
}

func CreateMountPoint(containerName string, imageName string) {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	exist, err := PathExists(mntUrl)
	if err != nil {
		log.Errorf("Error %v", err)
		return
	}
	if exist == false {
		if err := os.MkdirAll(mntUrl, 0777); err != nil {
			log.Errorf("fail to create directory %s : %v", mntUrl, err)
		}
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayer, containerName)
	tmpImageUrl := RootUrl + "/" + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("mount readonlyLayer and writeLayer error: %v", err)
	}
}

func DeleteWorkSpace(volumeStr string, containerName string) {
	valid, volumeUrls := ExtractVolumeParameter(volumeStr)
	if valid {
		DeleteMountPointWithVolume(volumeUrls, containerName)
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

func DeleteMountPointWithVolume(volumeUrls []string, containerName string) {
	/* unmount container volume. */
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntUrl + volumeUrls[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		log.Errorf("Unmount volume %s error : %v", containerUrl, err)
	}
	/* unmount total mount point volume. */
	if _, err := exec.Command("umount", mntUrl).CombinedOutput(); err != nil {
		log.Errorf("Unmount volume %s error : %v", mntUrl, err)
	}
	/* remove total mount point directory. */
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Remove volume %s error : %v", mntUrl, err)
	}
}

func DeleteMountPoint(containerName string) {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	_, err := exec.Command("umount", mntUrl).CombinedOutput()
	if err != nil {
		log.Errorf("Umount volume %s error: %v", mntUrl, err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Remove mount point %s error: %v", mntUrl, err)
	}
}

func DeleteWriteLayer(containerName string) {
	writeLayerUrl := fmt.Sprintf(WriteLayer, containerName)
	if err := os.RemoveAll(writeLayerUrl); err != nil {
		log.Errorf("remove writeLayer %s error: %v", writeLayerUrl, err)
	}
}
