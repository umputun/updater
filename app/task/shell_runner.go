package task

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/go-pkgz/lgr"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// ShellRunner executes commands with shell
type ShellRunner struct {
	BatchMode bool
}

// Run command in shell with provided logger
func (s *ShellRunner) Run(ctx context.Context, command string, logWriter io.Writer) error {
	if command == "" {
		return nil
	}

	command = strings.TrimSpace(command)
	suppressError := false
	if command[0] == '@' {
		command = command[1:]
		suppressError = true
		log.Printf("[DEBUG] suppress error for %s", command)
	}
	log.Printf("[INFO] execute %q", command)
	cmd := exec.CommandContext(ctx, "sh", "-c", command) // nolint
	if s.BatchMode {
		batchFile, err := s.prepBatch(command, suppressError)
		if err != nil {
			return fmt.Errorf("can't run comand in batch mode: %w", err)
		}
		defer func() {
			if e := os.Remove(batchFile); e != nil {
				log.Printf("[WARN] can't remove temp batch file %s, %v", batchFile, e)
			}
		}()
		cmd = exec.CommandContext(ctx, "sh", batchFile) // nolint
	}

	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	cmd.Stdin = os.Stdin
	log.Printf("[DEBUG] executing command: %s", command)

	err := cmd.Run()
	if err != nil {
		if suppressError {
			log.Printf("[WARN] suppressed error executing %s, %v", command, err)
			return nil
		}
		return err
	}

	return nil
}

func (s *ShellRunner) prepBatch(cmd string, suppressError bool) (batchFile string, err error) {
	var script []string
	script = append(script, "#!bin/sh")
	if !suppressError {
		script = append(script, "set -e")
	}
	script = append(script, strings.Split(cmd, "\n")...)
	fh, e := ioutil.TempFile("/tmp", "updater")
	if e != nil {
		return "", errors.Wrap(e, "failed to prep batch")
	}
	defer func() {
		errs := new(multierror.Error)
		fname := fh.Name()
		errs = multierror.Append(errs, fh.Sync())
		errs = multierror.Append(errs, fh.Close())
		errs = multierror.Append(errs, os.Chmod(fname, 0755)) //nolint
		if errs.ErrorOrNil() != nil {
			log.Printf("[WARN] can't properly close %s, %v", fname, errs.Error())
		}
	}()

	buff := bytes.NewBufferString(strings.Join(script, "\n"))
	_, err = io.Copy(fh, buff)
	return fh.Name(), errors.Wrapf(err, "failed to write to %s", fh.Name())
}
