package unsure

import (
	"flag"
	"fmt"
	llog "log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
)

var (
	crashTTL = flag.Duration("crash_ttl", time.Minute, "max duration before the app will crash (0 disables)")
	jsonLogs = flag.Bool("json_logs", false, "enable json jettison logs")
)

func crashDuration() (time.Duration, bool) {
	if crashTTL.Seconds() == 0 {
		return 0, false
	}
	nanos := rand.Int63n(crashTTL.Nanoseconds())
	return time.Nanosecond * time.Duration(nanos), true
}

func Bootstrap() {
	flag.Parse()
	initJettisonLogger()
	rand.Seed(time.Now().UnixNano())
}

func initJettisonLogger() {
	if *jsonLogs {
		return
	}
	l := llog.New(os.Stdout, "", 0)
	log.SetLogger(&logger{l})
}

func WaitForShutdown() {
	ch := make(chan os.Signal, 1)

	// crash before TTL
	go func() {
		d, ok := crashDuration()
		if !ok {
			return
		}
		time.Sleep(d)
		log.Info(nil, "app: The end is nigh")
		ch <- syscall.SIGKILL
	}()

	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	s := <-ch
	log.Info(nil, "app: Received OS signal", j.KV("signal", s))
	shutdown()
}

func ListenAndServeForever(addr string, handler http.Handler) {
	srv := &http.Server{Addr: addr, Handler: handler}
	RegisterShutdown(srv.Close)
	log.Info(nil, "app: Listening for HTTP requests", j.KV("address", addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		Fatal(err)
	}
}

func Fatal(err error) {
	if err == nil {
		return
	}
	log.Error(nil, errors.Wrap(err, "app fatal error"))
	shutdown()
}

type logger struct {
	*llog.Logger
}

func (logger *logger) Log(l log.Log) string {
	logger.Printf("[%s] %s: %s",
		strings.ToUpper(string(l.Level)),
		makePrefix(l.Source),
		makeMsg(l),
	)

	return ""
}

func makeMsg(l log.Log) string {
	var res strings.Builder
	res.WriteString(l.Message)
	if len(l.Parameters) == 0 {
		return res.String()
	}
	var pl []string
	for _, p := range l.Parameters {
		pl = append(pl, fmt.Sprintf("%s=%s", p.Key, p.Value))
	}
	res.WriteString("[")
	res.WriteString(strings.Join(pl, ","))
	res.WriteString("]")
	return res.String()
}

func makePrefix(source string) string {
	split := strings.Split(source, "/")
	var res []string
	for i, s := range split {
		if i < len(split)-2 {
			res = append(res, string([]rune(s)[0]))
		} else if i < len(split)-1 {
			res = append(res, strings.Split(s, ".")[0])
		}
		// else skip last element file.go:line
	}

	return strings.Join(res, ".")
}
