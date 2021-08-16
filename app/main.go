package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/go-pkgz/syncs"
	"github.com/umputun/go-flags"

	"github.com/umputun/updater/app/server"
	"github.com/umputun/updater/app/task"
)

var revision string

var opts struct {
	Config    string        `short:"f" long:"file" env:"CONF" default:"updater.yml" description:"config file"`
	Listen    string        `short:"l" long:"listen" env:"LISTEN" default:"localhost:8080" description:"listen on host:port"`
	SecretKey string        `short:"k" long:"key" env:"KEY" required:"true" description:"secret key"`
	Batch     bool          `short:"b" long:"batch" description:"batch mode for multi-line scripts"`
	Limit     int           `long:"limit"  default:"10" description:"limit how many concurrent update can be running"`
	TimeOut   time.Duration `long:"timeout"  default:"1m" description:"for how long update task can be running"`
	Dbg       bool          `long:"dbg" description:"show debug info"`
}

func main() {
	fmt.Printf("updater %s\n", revision)

	p := flags.NewParser(&opts, flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		p.WriteHelp(os.Stderr)
		os.Exit(2)
	}
	setupLog(opts.Dbg)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic:\n%v", x)
			panic(x)
		}

		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	conf, err := task.LoadConfig(opts.Config)
	if err != nil {
		log.Fatalf("[ERROR] can't load config %q, %v", opts.Config, err)
	}
	runner := &task.ShellRunner{BatchMode: opts.Batch, Limiter: syncs.NewSemaphore(opts.Limit), TimeOut: opts.TimeOut}

	srv := server.Rest{
		Listen:      opts.Listen,
		Version:     revision,
		SecretKey:   opts.SecretKey,
		Config:      conf,
		Runner:      runner,
		UpdateDelay: time.Second,
	}

	if err := srv.Run(ctx); err != nil {
		if !strings.Contains(err.Error(), "Server closed") {
			log.Fatalf("[ERROR] server failed, %v", err)
		}
	}
}

func setupLog(dbg bool) {
	if dbg {
		log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
		return
	}
	log.Setup(log.Msec, log.LevelBraces)
}
