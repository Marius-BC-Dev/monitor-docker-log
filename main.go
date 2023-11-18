package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-lark/lark"
	"github.com/urfave/cli/v2"
)

var (
	configFileFlag = cli.StringFlag{
		Name:     FlagCfg,
		Aliases:  []string{"c"},
		Usage:    "Configuration `FILE`",
		Required: true,
	}
)

func readContainerLog(cli *client.Client, container ContainerInfo, options types.ContainerLogsOptions, senderNotifyCH chan<- string) error {
	out, err := cli.ContainerLogs(context.Background(), container.ContainerID, options)
	if err != nil {
		return err
	}

	defer out.Close()

	// 从日志流中读取并处理日志行
	// 将日志流转换为扫描器
	scanner := bufio.NewScanner(out)

	for scanner.Scan() {
		line := scanner.Bytes()

		// 将[]byte切片转换为字符串输出
		// 因为docker前8个字符都是不可见字符，所以需要去掉，从第九个开始读取
		resultString := string(line)[8:]
		if strings.Contains(strings.ToLower(resultString), "error") || strings.Contains(strings.ToLower(resultString), "err") {
			senderNotifyCH <- resultString
		}
	}

	return nil
}

func notificationLark(bot *lark.Bot, recvNotifyCH <-chan string) {
	for msg := range recvNotifyCH {
		if _, err := bot.PostNotificationV2(lark.NewMsgBuffer(lark.MsgText).Text(msg).Build()); err != nil {
			fmt.Println(err)
		}
	}

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
}

func start(c *cli.Context) error {
	Init(LogConfig{
		Environment: EnvironmentDevelopment,
		Level:       "debug",
		Outputs:     []string{"stringdebug", "stderr"},
	})

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

	bot := lark.NewNotificationBot(cfg.HookUrl)
	ch := make(chan string, 10240)

	go func() {
		notificationLark(bot, ch)
	}()

	for i := 0; i < len(cfg.Containers); i++ {
		container := cfg.Containers[i]
		if container.HookUrl != "" {
			containerBot := lark.NewNotificationBot(cfg.HookUrl)
			containerCh := make(chan string, 10240)

			go func() {
				notificationLark(containerBot, containerCh)
			}()

			go func() {
				if err := readContainerLog(cli, container, options, containerCh); err != nil {
					fmt.Println(err)
				}
			}()
		} else {
			go func() {
				if err := readContainerLog(cli, container, options, ch); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}

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
