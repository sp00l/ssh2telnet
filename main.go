package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/ziutek/telnet"

	flags "github.com/jessevdk/go-flags"
)

type options struct {
	Addr           string `short:"a" long:"addr"   description:"Address to listen" default:":2222"`
	Target         string `short:"t" long:"target" description:"Telnet target address" default:"localhost:23"`
	HostKey        string `short:"k" long:"key"    description:"Path to the host key"`
	ClearScreen    bool   `short:"c" long:"clear"  description:"Send ANSI clear-screen sequence after a login"`
	AutoLogin      bool   `short:"l" long:"login"  description:"Enable auto login"`
	LoginPrompt    string `long:"login-prompt"     description:"Login prompt (default: \"login: \")" default:"login: " default-mask:"-"`
	PasswordPrompt string `long:"password-prompt"  description:"Password prompt (default: \"Password: \")" default:"Password: " default-mask:"-"`
}

func start(opts options) error {
	server := &ssh.Server{Addr: opts.Addr}

	if _, err := os.Stat(opts.HostKey); err == nil {
		hostKeyFile := ssh.HostKeyFile(opts.HostKey)
		server.SetOption(hostKeyFile)
		slog.Info("using host key", "hostkey", opts.HostKey)
	}

	if opts.AutoLogin {
		passwordAuth := ssh.PasswordAuth(func(ctx ssh.Context, s string) bool {
			ctx.SetValue("password", s)
			return true
		})
		server.SetOption(passwordAuth)
		slog.Info("using auto login", "loginprompt", opts.LoginPrompt, "passwordprompt", opts.PasswordPrompt)
	}

	server.Handle(func(s ssh.Session) {
		var username, password string
		username = s.User()

		l := slog.With(
			"id", s.Context().Value(ssh.ContextKeySessionID).(string)[:8],
			"user", username,
			"remote_ip", s.RemoteAddr().String(),
		)

		if opts.AutoLogin {
			password = s.Context().Value("password").(string)
		}

		l.Info("new session", "status", "started")

		_, _, isPty := s.Pty()
		if isPty {

			if opts.ClearScreen {
				io.WriteString(s, "\x1b[2J\x1b[H")
			}

			conn, err := telnet.Dial("tcp", opts.Target)
			if err != nil {
				l.Error("unable to connect to target", "status", "closed", "target", opts.Target)
				s.Exit(1)
				return
			}
			defer func() {
				conn.Close()
				l.Info("session closed", "status", "closed")
			}()

			if opts.AutoLogin {
				_, err = conn.ReadUntil(opts.LoginPrompt)
				conn.Write([]byte(fmt.Sprintf("%s\n", username)))
				_, err = conn.ReadUntil(opts.PasswordPrompt)
				conn.Write([]byte(fmt.Sprintf("%s\n", password)))
			}

			sigChan := make(chan struct{}, 1)
			go func() {
				_, _ = io.Copy(s, conn)
				sigChan <- struct{}{}
			}()
			go func() {
				_, _ = io.Copy(conn, s)
				sigChan <- struct{}{}
			}()

			<-sigChan
		} else {
			l.Error("no pty requested", "status", "closed")
			s.Exit(1)
		}
	})

	slog.Info("starting ssh server", "listen", opts.Addr, "target", opts.Target)
	return server.ListenAndServe()
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			os.Exit(0)
		}
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}

	if err := start(opts); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
