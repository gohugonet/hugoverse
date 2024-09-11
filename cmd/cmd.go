package cmd

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gohugonet/hugoverse/internal/interfaces/cli"
	"os"
)

func New() error {
	topLevel := flag.NewFlagSet("hugov", flag.ExitOnError)
	topLevel.Usage = func() {
		fmt.Println("Usage:\n  hugov [command]")
		fmt.Println("\nCommands:")
		fmt.Println("    serve:  start the headless CMS server")
		fmt.Println("  version:  show hugoverse command version")

		fmt.Println("\nExample:")
		fmt.Println("  hugov version")
	}

	err := topLevel.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	if topLevel.Parsed() {
		if len(topLevel.Args()) == 0 {
			topLevel.Usage()
			return errors.New("please specify a sub-command")
		}

		// 获取子命令及参数
		subCommand := topLevel.Args()[0]

		switch subCommand {
		case "version":
			versionCmd, err := cli.NewVersionCmd(topLevel)
			if err != nil {
				return err
			}
			if err := versionCmd.Run(); err != nil {
				return err
			}
		case "serve":
			serveCmd, err := cli.NewServeCmd(topLevel)
			if err != nil {
				return err
			}
			if err := serveCmd.Run(); err != nil {
				return err
			}
		case "demo":
			demoCmd, err := cli.NewDemoCmd(topLevel)
			if err != nil {
				return err
			}
			if err := demoCmd.Run(); err != nil {
				return err
			}
		case "build":
			openCmd, err := cli.NewBuildCmd(topLevel)
			if err != nil {
				return err
			}
			if err := openCmd.Run(); err != nil {
				return err
			}
		case "static":
			staticCmd, err := cli.NewStaticCmd(topLevel)
			if err != nil {
				return err
			}
			if err := staticCmd.Run(); err != nil {
				return err
			}

		default:
			topLevel.Usage()
			return errors.New("invalid sub-command")
		}
	}

	return nil
}
