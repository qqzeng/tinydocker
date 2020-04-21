package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os/exec"
)

func commitContainer(imageName string) {
	mntUrl := "/root/mnt/"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("commit image %s", imageTar)
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").
		CombinedOutput(); err != nil {
		log.Errorf("Create tar %s error : %v", imageTar, err)
	}
}
