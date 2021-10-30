package implementation

import (
	"errors"
	"github.com/appleboy/easyssh-proxy"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

var (
	errMissingHost          = errors.New("Error: missing server host")
	errMissingPasswordOrKey = errors.New("Error: can't connect without a private SSH key")
)

type (

	// Config for the plugin.
	Config struct {
		Key        string
		Passphrase string
		Username   string
		Host       string
		Port       int
		Timeout    time.Duration
		Script     []string
		Debug      bool
		Sync       bool
	}

	// Plugin structure
	Plugin struct {
		Config Config
	}
)

// Exec executes the plugin.
func (p Plugin) Exec() (string, error) {
	if len(p.Config.Host) == 0 {
		return "", errMissingHost
	}

	if len(p.Config.Key) == 0 {
		return "", errMissingPasswordOrKey
	}

	host, port := p.hostPort(p.Config.Host)

	// Create MakeConfig instance with remote username, server address and path to private key.
	ssh := &easyssh.MakeConfig{
		Server:     host,
		User:       p.Config.Username,
		Port:       port,
		Key:        p.Config.Key,
		Passphrase: p.Config.Passphrase,
		Timeout:    p.Config.Timeout,
		Proxy:      easyssh.DefaultConfig{},
	}

	log.Infof(host, "Running SSH command against host: %v, command: %v", host, p.Config.Script)

	outStr, errStr, _, err := ssh.Run(strings.Join(p.Config.Script, "\n"), p.Config.Timeout)
	if err != nil {
		return strings.Join([]string{outStr, errStr}, "\n"), err
	}

	return outStr, nil
}

func (p Plugin) hostPort(host string) (string, string) {
	hosts := strings.Split(host, ":")
	port := strconv.Itoa(p.Config.Port)
	if len(hosts) > 1 {
		host = hosts[0]
		port = hosts[1]
	}
	return host, port
}
