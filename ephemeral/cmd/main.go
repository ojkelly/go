// ephemeral
//
// watch for changes to your source code and restart your run command.
//
// It's called ephemeral, becuase we make sure the previous commend
// is completely gone before starting the new one.
//
// TODO: .gitignore/.ignore support
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v2"
	"ojkelly.dev/ephemeral"
)

func main() {
	var cmd string

	s := make(chan os.Signal, 1)
	errs := make(chan error, 1)
	close := make(chan struct{}, 1)

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	ctx, shutdown := context.WithCancel(context.Background())

	app := &cli.App{
		Name:                 "Ephemeral",
		HelpName:             "ephemeral",
		Usage:                "reload the app your developing when you change its source files",
		EnableBashCompletion: false, // TODO: make this true
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Owen Kelly",
				Email: "owen@ojkelly.dev",
			},
		},
		Metadata: map[string]interface{}{
			"Source": "https://github.com/ojkelly/go",
		},
		ArgsUsage: "*.go ../**/*.go ./**/**.ts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "command",
				Aliases:     []string{"c"},
				Usage:       "command to run: (example: \"go run cmd/main.go\" or \"node src/index.js\")",
				Required:    true,
				Destination: &cmd,
			},
		},
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			service, err := ephemeral.New(cmd, c.Args().Slice(), errs, close)
			if err != nil {
				return err
			}

			go service.Run(ctx)

			<-ctx.Done()
			log.Println("end of action")
			return nil
		},
	}

	go func() {
		err := app.Run(os.Args)
		if err != nil {
			log.Fatal(err)
		}

		for {
			err := <-errs
			log.Println("run err:", err)
		}
	}()

	select {
	case sig := <-s:
		fmt.Printf("\nExit with %s\n", sig.String())
	case <-ctx.Done():
		log.Println("top level done")
	}

	close <- struct{}{}

	log.Println("Waiting 500ms to shutdown")
	time.Sleep(time.Second / 2)

	// shutdown the app, the os will restart it
	shutdown()
	os.Exit(0)
	//
	defer log.Println("Done.")
}
