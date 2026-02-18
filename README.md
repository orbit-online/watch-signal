# watch-signal

Signal a process when paths change

```
Usage: watch-signal <signal> <pidfile> <paths> ... [flags]

Signal a process when paths change

Arguments:
  <signal>       The POSIX signal to send when a watched path changes
  <pidfile>      Path to the file containing the PID to send the signal to
  <paths> ...    Filesystem paths to watch for changes

Flags:
  -h, --help       Show context-sensitive help.
      --verbose    Turn on verbose logging
```
