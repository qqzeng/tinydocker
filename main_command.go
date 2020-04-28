package main

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"github.com/qqzeng/tinydocker/container"
	"github.com/qqzeng/tinydocker/network"
	log "github.com/Sirupsen/logrus"
	"os"
)

const (
	RootUrl = "/root/"
	MntUrl = "/root/mnt/"
	Usage = "tinydocker is a simple container runtime implementation for learning purpose."
	ENV_EXEC_PID = "tinydocker_pid"
	ENV_EXEC_COMMAND = "tinydocker_command"
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
		// ./tinydocker  run  -d  --name  containerl  -v  /root/froml:/tol  busybox  top
		imageName := cmdArray[0]
		cmdArray = cmdArray[1:]
		tty := context.Bool("it")
		volumeStr := context.String("v")
		detached := context.Bool("d")
		containerName := context.String("name")
		envSlice := context.StringSlice("e")
		network := context.String("net")
		portmapping := context.StringSlice("p")
		res := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		if tty == detached {
			return fmt.Errorf("option it and d can not be identical")
		}
		Run(tty, cmdArray, res, volumeStr, containerName, imageName, envSlice, network, portmapping)
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
		cli.StringSliceFlag{
			Name:  "e",
			Usage: "set environment variables",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name: "p",
			Usage: "port mapping",
		},
	},
}

var initCommand = cli.Command{
	Name:                   "init",
	Usage:                  "Init container process run user’s process in container.  Do not call it outside",
	Action: func(context *cli.Context) error {
		log.Info("init comes on")
		if context.NArg() < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get (0)
		err := container.RunContainerInitProcess(containerName)
		return err
	},
}

var commitCommand = cli.Command {
	Name:                   "commit",
	Usage:                  "Commit current running container into a image",
	Action: func(context *cli.Context) error {
		if context.NArg() < 2 {
			return fmt.Errorf("missing container name and image name")
		}
		containerName := context.Args().Get (0)
		imageName := context.Args().Get (1)
		commitContainer(containerName, imageName)
		return nil
	},

}

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
		containerName := context.Args().Get(0)
		LogContainer(containerName)
		return nil
	},
}

var execCommand = cli.Command{
	Name:                   "exec",
	Usage:                  "Execute a command in given container",
	Action: func(context *cli.Context) error {
		if os.Getenv(ENV_EXEC_PID) != "" {
			log.Infof("pid callback %v", os.Getgid())
			return nil
		}
		if context.NArg() < 2 {
			return fmt.Errorf("missing container name or command")
		}
		containerName := context.Args().Get (0)
		var comArray []string
		for _, arg := range context.Args().Tail() {
			comArray = append(comArray, arg)
		}
		ExecContainer(containerName, comArray)
		return nil
	},
}

var stopCommand = cli.Command{
	Name:                   "stop",
	Usage:                  "Stop a running container process",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get (0)
		StopContainer(containerName)
		return nil
	},
}

var removeCommand = cli.Command{
	Name:                   "rm",
	Usage:                  "Remove a unused container",
	Action: func(context *cli.Context) error {
		if context.NArg() < 1 {
			return fmt.Errorf("missing container name")
		}
		containerName := context.Args().Get (0)
		RemoveContainer(containerName)
		return nil
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "Container network commands",
	Subcommands: []cli.Command {
		{
			Name: "create",
			Usage: "Create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action:func(context *cli.Context) error {
				if context.NArg() < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				driver := context.String("driver")
				subnet := context.String("subnet")
				name := context.Args().Get (0)
				err := network.CreateNetwork(driver, subnet, name)
				return err
			},
		},
		{
			Name: "list",
			Usage: "Display container network list",
			Action:func(context *cli.Context) error {
				network.Init()
				network.ListNetwork()
				return nil
			},
		},
		{
			Name: "remove",
			Usage: "Remove a container network",
			Action:func(context *cli.Context) error {
				if context.NArg() < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				name := context.Args().Get (0)
				err := network.DeleteNetwork(name)
				return err
			},
		},
	},
}