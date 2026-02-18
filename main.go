package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
)

type Params struct {
	Signal  string   `required:"" arg:"" name:"signal" help:"The POSIX signal to send when a watched path changes (without the \"SIG\" prefix)" enum:"HUP,INT,QUIT,ILL,TRAP,ABRT,IOT,BUS,FPE,KILL,USR1,SEGV,USR2,PIPE,ALRM,TERM,CHLD,CONT,STOP,TSTP,TTIN,TTOU,URG,XCPU,XFSZ,VTALRM,WINCH,PROF,IO,SYS"`
	PidFile string   `required:"" arg:"" name:"pidfile" help:"Path to the file containing the PID to send the signal to"`
	Paths   []string `required:"" arg:"" name:"paths" help:"Filesystem paths to watch for changes"`
	Verbose bool     `help:"Turn on verbose logging"`
}

var params Params

var (
	signalMap = map[string]os.Signal{
		"HUP":    syscall.SIGHUP,
		"INT":    syscall.SIGINT,
		"QUIT":   syscall.SIGQUIT,
		"ILL":    syscall.SIGILL,
		"TRAP":   syscall.SIGTRAP,
		"ABRT":   syscall.SIGABRT,
		"IOT":    syscall.SIGIOT,
		"BUS":    syscall.SIGBUS,
		"FPE":    syscall.SIGFPE,
		"KILL":   syscall.SIGKILL,
		"USR1":   syscall.SIGUSR1,
		"SEGV":   syscall.SIGSEGV,
		"USR2":   syscall.SIGUSR2,
		"PIPE":   syscall.SIGPIPE,
		"ALRM":   syscall.SIGALRM,
		"TERM":   syscall.SIGTERM,
		"CHLD":   syscall.SIGCHLD,
		"CONT":   syscall.SIGCONT,
		"STOP":   syscall.SIGSTOP,
		"TSTP":   syscall.SIGTSTP,
		"TTIN":   syscall.SIGTTIN,
		"TTOU":   syscall.SIGTTOU,
		"URG":    syscall.SIGURG,
		"XCPU":   syscall.SIGXCPU,
		"XFSZ":   syscall.SIGXFSZ,
		"VTALRM": syscall.SIGVTALRM,
		"WINCH":  syscall.SIGWINCH,
		"PROF":   syscall.SIGPROF,
		"IO":     syscall.SIGIO,
		"SYS":    syscall.SIGSYS,
	}
)

func main() {
	kong.Parse(&params, kong.Name("watch-signal"), kong.Description("Signal a process when paths change"))
	slog.SetDefault(slog.Default())
	if params.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
	err := startWatchSignal(params)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func startWatchSignal(params Params) error {
	signal, found := signalMap[params.Signal]
	if !found {
		return fmt.Errorf("unknown signal \"%s\"", params.Signal)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create filesystem watcher: %w", err)
	}
	defer watcher.Close()
	for _, path := range params.Paths {
		err = watcher.Add(path)
		if err != nil {
			return fmt.Errorf("failed to watch path %s: %w", path, err)
		}
	}
	for {
		slog.Info("Startup completed")
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return fmt.Errorf("Watcher was closed")
			}
			slog.Debug("File changed", "path", event.Name)
			pidStr, err := os.ReadFile(params.PidFile)
			if err != nil {
				slog.Warn("Failed to read pidfile", "pidfile", params.PidFile, "err", err)
				break
			}
			pid, err := strconv.Atoi(strings.Trim(string(pidStr), " \n\t\r"))
			if err != nil {
				slog.Warn("Unable to parse pid as an integer", "pid", pidStr, "pidfile", params.PidFile, "err", err)
				break
			}
			proc := os.Process{Pid: pid}
			slog.Debug("Signalling", "signal", params.Signal, "pid", pid)
			proc.Signal(signal)
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("Watcher was closed")
			}
			slog.Warn("Error while watching for file changes", "err", err)
		}
	}
}
