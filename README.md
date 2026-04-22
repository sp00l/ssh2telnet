# ssh2telnet
Proxy ssh connection into telnet. Forked from [haccht/ssh2telnet](https://github.com/haccht/ssh2telnet).
Mine works in a Docker container and only with a fixed target.

## Options

```
$ docker run sp00l/ssh2telnet -h
Usage:
  ssh2telnet [OPTIONS]

Application Options:
  -a, --addr=            Address to listen (default: :2222)
  -t, --target=          Telnet target address (default: localhost:23)
  -k, --key=             Path to the host key
  -l, --login            Enable auto login
      --login-prompt=    Login prompt (default: "login: ")
      --password-prompt= Password prompt (default: "Password: ")

Help Options:
  -h, --help             Show this help message
```
## Basic Usage

Start a ssh server.

```
$ docker run -p 2222:2222 sp00l/ssh2telnet -t sometelnetserver.com:23
Starting ssh server on :2222
```

## Auto Login

ssh2telnet also comes with the auto login feature.
Start a ssh server with the `--login` option and specify `--login-prompt` and/or `--password-prompt` if necessary.

```
$ docker run -p 2222:2222 sp00l/ssh2telnet -t sometelnetserver.com:23 -l --login-prompt 'Username: '
Starting ssh server on :2222
```
