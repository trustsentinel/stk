// Package shell starts a login shell inside a pseudo-terminal, exposed as an
// io.ReadWriteCloser so the agent can pipe it over an encrypted session.
package shell

import (
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
)

// Session is a shell running in a PTY.
type Session struct {
	f   *os.File
	cmd *exec.Cmd
}

// Start launches shellPath (default /bin/sh) in a PTY.
func Start(shellPath string) (*Session, error) {
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	cmd := exec.Command(shellPath)
	cmd.Env = append(os.Environ(), "TERM=xterm")
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	return &Session{f: f, cmd: cmd}, nil
}

// Read reads terminal output.
func (s *Session) Read(b []byte) (int, error) { return s.f.Read(b) }

// Write writes terminal input.
func (s *Session) Write(b []byte) (int, error) { return s.f.Write(b) }

// Close closes the PTY and reaps the shell. It kills the process explicitly:
// closing the PTY master does not reliably terminate the shell on every platform,
// and without this the agent could hang in Wait() and never serve again.
func (s *Session) Close() error {
	_ = s.f.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	_ = s.cmd.Wait()
	return nil
}

var _ io.ReadWriteCloser = (*Session)(nil)
