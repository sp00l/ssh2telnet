package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gliderlabs/ssh"
	"github.com/ziutek/telnet"

	flags "github.com/jessevdk/go-flags"
)

type options struct {
	Addr           string `short:"a" long:"addr"   description:"Address to listen" default:":2222"`
	Target         string `short:"t" long:"target" description:"Telnet target address" default:"localhost:23"`
	HostKey        string `short:"k" long:"key"    description:"Path to the host key"`
	AutoLogin      bool   `short:"l" long:"login"  description:"Enable auto login"`
	LoginPrompt    string `long:"login-prompt"     description:"Login prompt (default: \"login: \")" default:"login: " default-mask:"-"`
	PasswordPrompt string `long:"password-prompt"  description:"Password prompt (default: \"Password: \")" default:"Password: " default-mask:"-"`
}

func start(opts options) error {
	server := &ssh.Server{Addr: opts.Addr}

	if _, err := os.Stat(opts.HostKey); err == nil {
		hostKeyFile := ssh.HostKeyFile(opts.HostKey)
		server.SetOption(hostKeyFile)
	}

	if opts.AutoLogin {
		passwordAuth := ssh.PasswordAuth(func(ctx ssh.Context, s string) bool {
			ctx.SetValue("password", s)
			return true
		})
		server.SetOption(passwordAuth)
	}

	server.Handle(func(s ssh.Session) {
		var username, password string
		if opts.AutoLogin {
			username = s.User()
			password = s.Context().Value("password").(string)
		}

		_, _, isPty := s.Pty()
		if isPty {

			fmt.Printf("Connecting to %s\n", opts.Target)

			conn, err := telnet.Dial("tcp", opts.Target)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Unable to connect to %s.\n", opts.Target)
				s.Exit(1)
				return
			}
			defer func() {
				conn.Close()
				fmt.Printf("Connection to %s closed\n", opts.Target)
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
			fmt.Fprintf(os.Stderr, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	fmt.Printf("Starting ssh server on %s\n", opts.Addr)
	return server.ListenAndServe()
}

func main() {
	var opts options
	if _, err := flags.Parse(&opts); err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			os.Exit(0)
		}
		log.Fatal(err)
	}

	if err := start(opts); err != nil {
		log.Fatal(err)
	}
}
