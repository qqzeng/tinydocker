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
		detached := context.Bool("d")
		containerName := context.String("name")
		res := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		if tty && detached {
			return fmt.Errorf("option it and d can not be both provided")
		}
		Run(tty, cmdArray, res, volumeStr, containerName)
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
		/* when testing detaching container, do not use `./tinydocker run -d top`,
		use `./tinydocker run -d top -b [-n 10]` instead*/
		cli.BoolFlag{
			Name:  "d",
			Usage: "detach container",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "container name",
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

var commitCommand = cli.Command {
	Name:                   "commit",
	Usage:                  "Commit current running container into a image",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("missing container command")
		}
		imageName := context.Args().Get (0)
		commitContainer(imageName)
		return nil
	},

}

/* TODO: `./tinydocker ps` does not update the status of container process.   */
var listCommand = cli.Command{
	Name:                   "ps",
	Usage:                  "List all containers in any status",
	Action: func(context *cli.Context) error {
		ListContainers()
		return nil
	},
}

var logCommand = cli.Command{
	Name:                   "logs",
	Usage:                  "Print logs of a container",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("missing container command")
		}
		containerName := context.Args().Get (0)
		LogContainer(containerName)
		return nil
	},
}