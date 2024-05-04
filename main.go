package main

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"miniDocker/subsystems"
	"os"
)

const usage = `a toy docker as a simple container runtime implementation`

func main() {
	app := cli.NewApp()
	app.Name = "minidocker"
	app.Usage = usage
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	setUpCommands(app)

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

// 设置CLI命令
func setUpCommands(app *cli.App) {
	// miniDocker run
	runCommand := cli.Command{
		Name:  "run",
		Usage: `Create a container`,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "ti",
				Usage: "run with terminal",
			},
			cli.StringFlag{
				Name:  "mem",
				Usage: "set memory limit",
			},
			cli.StringFlag{
				Name:  "cpushare",
				Usage: "set cpushare limit",
			},
			cli.StringFlag{
				Name:  "cpuset",
				Usage: "set cpuset limit",
			},
			cli.StringFlag{
				Name:  "v",
				Usage: "volume",
			},
			cli.StringFlag{
				Name:  "n",
				Usage: "set container name",
			},
		},
		Action: func(context *cli.Context) error {
			if len(context.Args()) < 1 {
				return fmt.Errorf("usage: minidocker run [cmd]")
			}
			var cmd []string
			for _, arg := range context.Args() {
				cmd = append(cmd, arg)
			}
			tty := context.Bool("ti")
			res := &subsystems.ResourceConfig{
				MemoryLimit: context.String("mem"),
				CPUSet:      context.String("cpuset"),
				CPUShare:    context.String("cpushare"),
			}
			volume := context.String("v")
			name := context.String("n")
			if name == "" {
				name = uuid.NewV4().String()
			}
			RunHandler(tty, cmd, res, volume, name)
			return nil
		},
	}

	initCommand := cli.Command{
		Name:  "init",
		Usage: `an inner command for container initiation`,
		Action: func(context *cli.Context) error {
			err := RunContainerInitProcess()
			return err
		},
	}

	commitCommand := cli.Command{
		Name:  "commit",
		Usage: "commit a container into image",
		Action: func(context *cli.Context) error {
			if len(context.Args()) != 1 {
				return fmt.Errorf("minidocker commit [containerName]")
			}
			imageName := context.Args().Get(0)
			commitContainer(imageName)
			return nil
		},
	}

	listCommand := cli.Command{
		Name:  "ps",
		Usage: "list containers",
		Action: func(ctx *cli.Context) error {
			listContainers()
			return nil
		},
	}

	app.Commands = []cli.Command{
		runCommand,
		initCommand,
		commitCommand,
		listCommand,
	}
}
