package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// ShellRunner executes commands with shell
type ShellRunner struct {
	BatchMode bool
	Limiter   sync.Locker
	TimeOut   time.Duration
}

// Run command in shell with provided logger
func (s *ShellRunner) Run(ctx context.Context, command string, logWriter io.Writer)(string, error) {
	if command == "" {
		return "", nil
	}

	if s.Limiter != nil {
		s.Limiter.Lock()
		defer s.Limiter.Unlock()
	}

	command = strings.TrimSpace(command)
	if s.BatchMode {
		batchFile, err := s.prepBatch(command)
		if err != nil {
			return "", fmt.Errorf("can't prepare batch: %w", err)
		}
		return "", s.runBatch(batchFile, logWriter, s.TimeOut)
	}

	execCmd := func(command string) (string, error) {
		log.Printf("[INFO] execute %q", command)
		var suppressError bool
		if command[0] == '@' {
			command = command[1:]
			suppressError = true
			log.Printf("[DEBUG] suppress error for %s", command)
		}
		cmd := exec.CommandContext(ctx, "sh", "-c", command) // nolint
		var outb bytes.Buffer
		cmd.Stdout = &outb
		cmd.Stderr = logWriter
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			if suppressError {
				log.Printf("[WARN] suppressed error executing %q, %v", command, err)
				return outb.String(), nil
			}
			return "", fmt.Errorf("failed to execute %s: %w", command, err)
		}
		res := outb.String()
		log.Printf("RES: ");
		log.Printf(res);
		return res, nil
	}
    var msg string
	for _, c := range strings.Split(command, "\n") {
		if c = strings.TrimSpace(c); c == "" {
			continue
		}
		if msg, err := execCmd(c); err != nil {
			return msg, err
		}
	}

	return msg, nil
}

func (s *ShellRunner) runBatch(batchFile string, logWriter io.Writer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer func() {
		cancel()
		if e := os.Remove(batchFile); e != nil {
			log.Printf("[WARN] can't remove temp batch file %s, %v", batchFile, e)
		}
	}()
	cmd := exec.CommandContext(ctx, "sh", batchFile) // nolint
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Stdin = os.Stdin
	log.Printf("[DEBUG] executing batch commands: %s", batchFile)

	return cmd.Run()
}

func (s *ShellRunner) prepBatch(cmd string) (batchFile string, err error) {
	var script []string
	script = append(script, "#!bin/sh")
	script = append(script, strings.Split(cmd, "\n")...)
	fh, e := os.CreateTemp("/tmp", "updater")
	if e != nil {
		return "", errors.Wrap(e, "failed to prep batch")
	}
	defer func() {
		errs := new(multierror.Error)
		fname := fh.Name()
		errs = multierror.Append(errs, fh.Sync())
		errs = multierror.Append(errs, fh.Close())
		errs = multierror.Append(errs, os.Chmod(fname, 0755)) // nolint
		if errs.ErrorOrNil() != nil {
			log.Printf("[WARN] can't properly close %s, %v", fname, errs.Error())
		}
	}()

	buff := bytes.NewBufferString(strings.Join(script, "\n"))
	_, err = io.Copy(fh, buff)
	return fh.Name(), errors.Wrapf(err, "failed to write to %s", fh.Name())
}
