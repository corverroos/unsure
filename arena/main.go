// Command arena executes a match by running the engine and player implementations at a difficulty
// level (increasing fate_p and decreasing crash_ttl) storing output.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/corverroos/unsure"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

var (
	initCmd     = flag.String("init", "", "init command (build script?) to run before starting match")
	engineCmd   = flag.String("engine", "engine", "command the run the engine")
	playerCmd   = flag.String("player", "player", "command the run a player")
	playerFlags = flag.String("player_flags", "--index=%d", "pipe separated player flags ($INDEX,$COUNT will be populated)")
	levelf      = flag.Int("level", 0, "difficulty level: [0-9]")

	levels = map[Level]struct {
		Rounds   int
		Players  int
		FateP    float64
		CrashTTL int // seconds
	}{
		// 0 fate_p 0 crash_ttl
		0: {1, 1, 0, 0},
		1: {5, 1, 0, 0},
		2: {10, 1, 0, 0},
		3: {1, 2, 0, 0},
		4: {5, 2, 0, 0},
		5: {10, 2, 0, 0},
		6: {1, 3, 0, 0},
		7: {5, 3, 0, 0},
		8: {10, 3, 0, 0},
	}
)

type Level int

func (l Level) UnsureFlags() []string {
	return []string{
		fmt.Sprintf("--fate_p=%0.2f", levels[l].FateP),
		fmt.Sprintf("--crash_ttl=%ds", levels[l].CrashTTL),
	}
}
func (l Level) EngineFlags() []string {
	return []string{
		fmt.Sprintf("--rounds=%d", levels[l].Rounds),
	}
}

func (l Level) Players() int {
	return levels[l].Players
}

func main() {
	unsure.Bootstrap()
	unsure.Surify()

	l := Level(*levelf)
	ctx, done := lifeCtx()

	out := new(output)

	if *initCmd != "" {
		log.Info(ctx, "Running init command", j.KV("cmd", *initCmd))
		out, err := exec.Command(*initCmd).CombinedOutput()
		if err != nil {
			log.Error(ctx, errors.Wrap(err, "prep command error", j.KV("out", string(out))))
			return
		}
	}

	fork(func() {
		runEngine(ctx, done, l, out.make(-1))
	})

	// Allow server to start up
	time.Sleep(time.Second)

	for i := 0; i < l.Players(); i++ {
		ii := i
		fork(func() {
			runPlayer(ctx, done, l, out.make(ii), ii)
		})
	}

	unsure.WaitForShutdown()
}

type output struct {
	writers map[int]io.WriteCloser
	mu      sync.Mutex
}

func (o *output) make(index int) io.WriteCloser {
	o.mu.Lock()
	defer o.mu.Unlock()

	out := getOut(index)

	if o.writers == nil {
		o.writers = make(map[int]io.WriteCloser)
	}
	o.writers[index] = out
	return out
}

// fork starts a go routine calling fn, it also registers a waitgroup blocking shutdown until the function exits.
func fork(fn func()) {
	var wg sync.WaitGroup
	wg.Add(1)

	unsure.RegisterNoErr(wg.Wait)
	go func() {
		fn()
		wg.Done()
	}()
}

func getOut(index int) io.WriteCloser {
	var name string
	if index == -1 {
		name = "engine.log"
	} else {
		name = fmt.Sprintf("player%d.log", index)
	}
	const dir = "/tmp/arena"
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		unsure.Fatal(err)
	}
	f, err := os.Create(path.Join(dir, name))
	if err != nil {
		unsure.Fatal(err)
	}
	return f
}

func lifeCtx() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	unsure.RegisterNoErr(cancel)
	go func() {
		<-ctx.Done()
		unsure.Fatal(errors.Wrap(ctx.Err(), "life context done"))
	}()
	return ctx, cancel
}

func runPlayer(ctx context.Context, done func(), l Level, out io.WriteCloser, n int) {
	defer done()
	defer out.Close()

	jopt := j.KV("index", n)

	flagStr := *playerFlags
	flagStr = strings.Replace(flagStr, "$INDEX", strconv.Itoa(n), 1)
	flagStr = strings.Replace(flagStr, "$COUNT", strconv.Itoa(l.Players()), 1)

	first := true
	for {
		flags := l.UnsureFlags()
		flags = append(flags, "--engine_address=127.0.0.1:12048")
		if first {
			flags = append(flags, "--db_recreate")
			first = false
		}
		if flagStr != "" {
			flags = append(flags, strings.Split(flagStr, "|")...)
		}

		log.Info(ctx, "starting player", j.KV("flags", flags), jopt)
		cmd := exec.Command(*playerCmd, flags...)
		cmd.Stdout = out
		cmd.Stderr = out
		err := cmd.Start()
		if err != nil {
			log.Error(ctx, errors.Wrap(err, "start player error", jopt))
			return
		}

		exit := make(chan error, 1)
		go func() {
			exit <- cmd.Wait()
		}()

		select {
		case err := <-exit:
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "run player error", jopt))
			} else {
				log.Info(ctx, "player completed", jopt)
			}
			return
			// TODO(corver): Implement restarts

		case <-ctx.Done():
			err := cmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "error terminating player", jopt))
			}

			err = <-exit
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "error after terminating engine"))
			}
			return
		}
	}
}

func runEngine(ctx context.Context, done func(), l Level, out io.WriteCloser) {
	defer done()
	defer out.Close()

	first := true
	for {
		flags := l.UnsureFlags()
		flags = append(flags, l.EngineFlags()...)
		if first {
			flags = append(flags, "--db_recreate")
			first = false
		}

		log.Info(ctx, "starting engine", j.KV("flags", flags))
		cmd := exec.Command(*engineCmd, flags...)
		cmd.Stdout = out
		cmd.Stderr = out
		err := cmd.Start()
		if err != nil {
			log.Error(ctx, errors.Wrap(err, "start engine error"))
			return
		}

		exit := make(chan error, 1)
		go func() {
			exit <- cmd.Wait()
		}()

		select {
		case err := <-exit:
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "run engine error"))
			} else {
				log.Info(ctx, "engine completed")
			}
			return
			// TODO(corver): Implement restarts

		case <-ctx.Done():
			err := cmd.Process.Signal(syscall.SIGTERM)
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "error terminating engine"))
			}

			err = <-exit
			if err != nil {
				log.Error(ctx, errors.Wrap(err, "error after terminating engine"))
			}
			return
		}
	}
}
