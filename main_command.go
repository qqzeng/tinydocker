package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"github.com/qqzeng/tinydocker/container"
	log "github.com/Sirupsen/logrus"
)

const (
	RootUrl = "/root/"
	MntUrl = "/root/mnt/"
	Usage = "tinydocker is a simple container runtime implementation for learning purpose."
)

var commitCommand = cli.Command {
	Name:                   "commit",
	Usage:                  "Commit current running container into a image",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container command")
		}
		imageName := context.Args().Get (0)
		commitContainer(imageName)
		return nil
	},
}

var runCommand = cli.Command {
	Name:                   "run",
	Usage:                  "Create a container with namespace and cgroups limit tinydocker run -it [command]",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("Missing container command")
		}
		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}
		cmdArray = cmdArray[0:]
		tty := context.Bool("it")
		volumeStr := context.String("v")
		res := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		Run(tty, cmdArray, res, volumeStr)
		return nil
	},
	Flags: [] cli.Flag {
		cli.BoolFlag{
			Name: "it",
			Usage: "keep STDIN open and enable tty",
		},
		cli.StringFlag{
			Name:  "m",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpushare",
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "v",
			Usage: "volume",
		},
	},
}


var initCommand = cli.Command{
	Name:                   "init",
	Usage:                  "Init container process run user’s process in container.  Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Info("init comes on")
		err := container.RunContainerInitProcess()
		return err
	},
}

