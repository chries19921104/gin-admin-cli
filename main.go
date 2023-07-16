package main

import (
	"log"
	"os"

	"github.com/chries19921104/gin-admin-cli/v1/cmd"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "gin-admin-cli"
	app.Description = "gin-admin 辅助工具，提供创建项目、快速生成功能模块的功能"
	app.Version = "5.1.0"
	app.Commands = []cli.Command{
		cmd.NewCommand(),
		cmd.GenerateCommand(),
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalf(err.Error())
	}
}
