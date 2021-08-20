package ephemeral

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Service struct {
	Watcher *fsnotify.Watcher
	closed  chan struct{}

	command string
	filters []string
	errs    chan error
	run     chan struct{}
	restart chan struct{}
	// cmd     *exec.Cmd
	killProcess chan error

	cmdStart     time.Time
	cmdMtx       sync.Mutex
	isRunning    bool
	isRunningMtx sync.Mutex
}

func New(cmd string, filters []string, errs chan error, closed chan struct{}) (*Service, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &Service{
		Watcher:     watcher,
		command:     cmd,
		filters:     filters,
		run:         make(chan struct{}, 1),
		errs:        errs,
		closed:      closed,
		killProcess: make(chan error, 1),
		restart:     make(chan struct{}, 1),
	}, nil
}

// Run to watch files and reload when they change
// on reload, also see if theres new files to watch
func (s *Service) Run(ctx context.Context) {
	defer s.Watcher.Close()

	err := s.findFilesToWatch()
	if err != nil {
		s.errs <- err
	}

	go s.throttle(ctx)
	// go s.throttleRun(ctx)

	// go s.runCommand(ctx)

	// https://github.com/fsnotify/fsnotify/issues/372
	go s.watch(ctx)
	s.run <- struct{}{}

	<-ctx.Done()
}

func (s *Service) watch(ctx context.Context) {
	for {
		select {
		case event, ok := <-s.Watcher.Events:
			if !ok {
				s.errs <- fmt.Errorf("error: bad watch event")
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
				s.restart <- struct{}{}
				log.Println("watch after s.run")
			}
		case err, ok := <-s.Watcher.Errors:
			if !ok {
				s.errs <- err
			}
			log.Println("error:", err)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) throttle(ctx context.Context) {

	var cmdCtx context.Context
	var kill context.CancelFunc
	cmdCtx, kill = context.WithCancel(context.Background())

	ready := true

	for {
		select {
		case <-s.restart:
			waitTime := s.cmdStart.Add(time.Second * 1)

			if time.Now().After(waitTime) {
				log.Println("time to kill then run", time.Now().After(waitTime))
				// s.killProcess <- fmt.Errorf("file changed")
				kill()

				fmt.Println("is ready", ready)
				if ready {
					cmdCtx, kill = context.WithCancel(context.Background())
					go s.runCommand(cmdCtx, kill)
					ready = false
				}
			}
		case <-s.run:
			// if ready {
			// 	cmdCtx, kill = context.WithCancel(context.Background())
			// 	go s.runCommand(cmdCtx, kill)
			// 	ready = false
			// }
		case <-cmdCtx.Done():
			ready = true
		case <-ctx.Done():
			kill()
		case <-s.closed:
			kill()
			return
		}
	}
}

// func (s *Service) throttleRun(ctx context.Context) {
// 	for {
// 		select {
// 		case <-time.After(time.Second):
// 			s.isRunningMtx.Lock()

// 			if !s.isRunning {
// 				log.Println("time to run", s.isRunning)

// 				go s.runCommand(ctx)
// 			}

// 			s.isRunningMtx.Unlock()

// 		case <-s.closed:
// 			return
// 		}
// 	}
// }

func (s *Service) findFilesToWatch() (err error) {
	for _, f := range s.filters {
		matches, err := filepath.Glob(f)
		if err != nil {
			return err
		}

		for _, match := range matches {
			err = s.Watcher.Add(match)
			if err != nil {
				return err
			}
		}
	}

	return err
}

func (s *Service) runCommand(ctx context.Context, kill context.CancelFunc) {
	log.Println("runCommand")

	// s.cmdMtx.Lock()
	// log.Println("runCommand past lock", s.isRunning)

	// defer s.cmdMtx.Unlock()

	// s.isRunningMtx.Lock()
	// s.isRunning = true
	// s.isRunningMtx.Unlock()

	s.cmdStart = time.Now()
	log.Println("runCommand set start time")
	cmdCtx := context.Context(ctx)

	cmd := exec.CommandContext(cmdCtx, "/bin/sh", "-c", s.command)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	log.Println("runCommand make cmd")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Println("runCommand before goru")

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	pid := cmd.Process.Pid

	go func() {
		e := cmd.Wait()
		log.Println("runCommand e", e)
		log.Println("runCommand  in goru")

		// time.Sleep(time.Second)
		s.killProcess <- e
		log.Println("runCommand  in goru after sned")

	}()
	// go func() {
	// 	err := cmd.Wait()
	// 	log.Println("runCommand goru", err)

	// 	s.killProcess <- err
	// }()

	log.Println("runCommand  after goru")

	select {
	case err := <-s.killProcess:
		log.Println("runCommand killProcess", cmd, err)
	case <-s.closed:
		log.Println("runCommand closed", cmd)
	case <-ctx.Done():
		log.Println("runCommand closed  ctx", cmd)
	}

	log.Println("runCommand after select")

	// kill()
	// cmd.Process.Kill()

	// if cmd.ProcessState == nil && cmd.Process != nil {
	fmt.Println("try killing", pid)

	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		fmt.Println("err killing", pid, err)
		s.errs <- err
	}

	// }

	// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
	// if cmd.ProcessState == nil {
	// 	log.Println("killProcess:", cmd.Process.Pid)
	// 	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	// 	if err != nil {
	// 		s.errs <- err
	// 	}
	// 	// Wait releases any resources associated with the Process.
	// 	_, _ = cmd.Process.Wait()
	// }
	// s.isRunningMtx.Lock()
	// s.isRunning = false
	// log.Println("end runCommand", s.isRunning)
	// s.isRunningMtx.Unlock()

	//  run again
	time.Sleep(time.Second * 5)
	// s.run <- struct{}{}
}
