package update

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// Values are positional arguments, never interpolated into shell source. The
// helper has a bounded wait and starts only after the old process releases its
// single-instance lock. It must outlive the request and the outgoing process.
const appImageRestartScript = `pid="$1"; shift
remaining=60
while kill -0 "$pid" 2>/dev/null; do
  [ "$remaining" -gt 0 ] || exit 1
  sleep 1
  remaining=$((remaining - 1))
done
exec "$@"
`

func startAppImageAfterExit(target string, args []string) error {
	argv := append([]string{"-c", appImageRestartScript, "mrrss-update", strconv.Itoa(os.Getpid()), target}, args...)
	cmd := exec.Command("/bin/sh", argv...)
	cmd.Dir = filepath.Dir(target)
	cmd.Env = appImageRestartEnv(os.Environ(), os.Getenv("APPDIR"))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func appImageRestartEnv(env []string, mount string) []string {
	result := make([]string, 0, len(env))
	for _, entry := range env {
		key, value, _ := strings.Cut(entry, "=")
		switch key {
		case "APPIMAGE", "APPDIR", "ARGV0":
			continue
		case "PATH", "LD_LIBRARY_PATH":
			if mount != "" {
				paths := []string{}
				for _, path := range filepath.SplitList(value) {
					if path != mount && !strings.HasPrefix(path, mount+"/") {
						paths = append(paths, path)
					}
				}
				entry = key + "=" + strings.Join(paths, ":")
			}
		}
		result = append(result, entry)
	}
	return result
}
