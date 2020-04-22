package main

import (
	"encoding/json"
	"fmt"
	"github.com/qqzeng/tinydocker/container"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

/* TODO: `./tinydocker ps` does not update the status of container process.   */
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
