package main

import (
	//"./cgroups/subsystems"
	"github.com/qqzeng/tinydocker/cgroups/subsystems"
	"strings"

	//"./cgroups"
	"github.com/qqzeng/tinydocker/cgroups"
	//"./container"
	"github.com/qqzeng/tinydocker/container"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = "tinydocker is a simple container runtime implementation for learning purpose."

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
		res := &subsystems.ResourceConfig{
			MemoryLimit: context.String("m"),
			CpuSet:      context.String("cpuset"),
			CpuShare:    context.String("cpushare"),
		}
		Run(tty, cmdArray, res)
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

func main() {
	app := cli.NewApp()
	app.Name  = "tinydocker"
	app.Usage  =  usage

	app.Commands = []cli.Command {
		initCommand,
		runCommand,
	}

	app.Before = func(context *cli.Context) error {
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func Run(tty bool, comArray []string, res *subsystems.ResourceConfig) {
	parent, wp := container.NewParentProcess(tty)
	if parent == nil {
		log.Error("new parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	sendInitCommand(comArray, wp)

	cgroupManager := cgroups.NewCgroupManager("tinydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	parent.Wait()
	os.Exit(-1)
}

func sendInitCommand(comArray []string, wp *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	wp.WriteString(command)
	wp.Close()
}
