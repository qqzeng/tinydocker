package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/qqzeng/tinydocker/container"
	"os/exec"
)

func commitContainer(containerName string, imageName string) {
	mntUrl  :=  fmt.Sprintf (container.MntUrl,  containerName)
	mntUrl += "/"
	imageTar := container.RootUrl  + "/" +  imageName  + ".tar"
	log.Infof("commit image %s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").
		CombinedOutput(); err != nil {
		log.Errorf("Create tar %s error : %v", imageTar, err)
	}
}
