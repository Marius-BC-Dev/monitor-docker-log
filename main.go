package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/urfave/cli/v2"
)

func readContainerLogs(containerID string) error {
	cli, err := client.NewClientWithOpts(client.WithVersion("1.41")) // 根据您的Docker版本选择合适的API版本
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "50", // 获取最后50行日志，根据需求调整
	}

	out, err := cli.ContainerLogs(context.Background(), containerID, options)
	if err != nil {
		return err
	}

	defer out.Close()

	// 从日志流中读取并处理日志行
	// 将日志流转换为扫描器
	scanner := bufio.NewScanner(out)
	//不断读取
	for scanner.Scan() {
		line := scanner.Bytes()

		// 将[]byte切片转换为字符串输出
		// 因为docker前8个字符都是不可见字符，所以需要去掉，从第九个开始读取
		resultString := string(line)[8:]
		// fmt.PrintLn(resultString)
		fmt.Println(resultString)
	}

	return nil
}

var (
	configFileFlag = cli.StringFlag{
		Name:     FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration `FILE`",
		Required: true,
	}
)

func readContainerLog(cli *client.Client, container ContainerInfo, options types.ContainerLogsOptions) error {
	out, err := cli.ContainerLogs(context.Background(), container.ContainerID, options)
	if err != nil {
		return err
	}

	defer out.Close()

	// 从日志流中读取并处理日志行
	// 将日志流转换为扫描器
	scanner := bufio.NewScanner(out)
	//不断读取
	for scanner.Scan() {
		line := scanner.Bytes()

		// 将[]byte切片转换为字符串输出
		// 因为docker前8个字符都是不可见字符，所以需要去掉，从第九个开始读取
		resultString := string(line)[8:]
		// fmt.PrintLn(resultString)
		fmt.Println(resultString)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "monitor docker log"
	app.Flags = []cli.Flag{
		&configFileFlag,
	}

	app.Action = start

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// containerID := "3a09b0e99be9" // 将此处替换为要读取日志的容器ID
	// err := readContainerLogs(containerID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
}

func start(c *cli.Context) error {
	cfg, err := load(c)
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.WithVersion(cfg.DockerVersion))
	if err != nil {
		return err
	}

	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       cfg.Tail, // 获取最后50行日志，根据需求调整
	}

	go func() {
		if err := readContainerLog(cli, cfg.Containers[0], options); err != nil {

		}
	}()

	// var cancelFuncs []context.CancelFunc

	waitSignal()
	return nil
}

// func setupLog(c log.Config) {
// 	log.Init(c)
// }

// func waitSignal(cancelFuncs []context.CancelFunc) {
func waitSignal() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			fmt.Println("terminating application gracefully...")

			exitStatus := 0
			// for _, cancel := range cancelFuncs {
			// 	cancel()
			// }
			os.Exit(exitStatus)
		}
	}
}
