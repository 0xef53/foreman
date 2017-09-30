package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/gcfg.v1"
)

// CommonParams represents the common parameters from the Foreman config file.
type CommonParams struct {
	ClientID string `gcfg:"client-id"`
	Servers  string
	Channel  string
}

// TopicParams represents the topic parameters such as worker script name,
// number of concurrent processes, notify hooks etc.
type TopicParams struct {
	Servers      string
	Channel      string
	Workdir      string
	Cmd          string
	Concurrency  int
	MaxAttempts  int    `gcfg:"max-attempts"`
	NotifyStart  string `gcfg:"notify-start"`
	NotifyFinish string `gcfg:"notify-finish"`
	NotifyFault  string `gcfg:"notify-fault"`
}

// ForemanConfig represents the Foreman configuration.
type ForemanConfig struct {
	Common        CommonParams
	Topic         map[string]*TopicParams
	Default_Topic TopicParams
}

// NewConfig reads and parses the configuration file and returns
// a new instance of ForemanConfig on success.
func NewConfig(p string) (*ForemanConfig, error) {
	cfg := ForemanConfig{
		Default_Topic: TopicParams{
			MaxAttempts: 1,
			Concurrency: 1,
		},
	}

	err := gcfg.ReadFileInto(&cfg, p)
	if err != nil {
		return nil, NewConfigError("failed to parse file:", err)
	}

	if len(cfg.Topic) == 0 {
		return nil, NewConfigError("no one topic defined")
	}

	if len(cfg.Common.Channel) == 0 {
		cfg.Common.Channel = "foreman"
	}

	serverRequired := false

	for name, opts := range cfg.Topic {
		if len(opts.Servers) == 0 {
			opts.Servers = cfg.Common.Servers
			serverRequired = true
		}
		if len(opts.Channel) == 0 {
			opts.Channel = cfg.Common.Channel
		}

		opts.Workdir = filepath.Join(opts.Workdir)
		opts.Cmd = filepath.Join(opts.Cmd)

		opts.NotifyStart = filepath.Join(opts.NotifyStart)
		opts.NotifyFinish = filepath.Join(opts.NotifyFinish)
		opts.NotifyFault = filepath.Join(opts.NotifyFault)

		if len(opts.Cmd) == 0 {
			return nil, NewConfigError("command is not defined for topic", name)
		}
		if !strings.HasPrefix(opts.Cmd, "/") {
			opts.Cmd = filepath.Join(opts.Workdir, opts.Cmd)
		}
		if len(opts.NotifyStart) > 0 && !strings.HasPrefix(opts.NotifyStart, "/") {
			opts.NotifyStart = filepath.Join(opts.Workdir, opts.NotifyStart)
		}
		if len(opts.NotifyFinish) > 0 && !strings.HasPrefix(opts.NotifyFinish, "/") {
			opts.NotifyFinish = filepath.Join(opts.Workdir, opts.NotifyFinish)
		}
		if len(opts.NotifyFault) > 0 && !strings.HasPrefix(opts.NotifyFault, "/") {
			opts.NotifyFault = filepath.Join(opts.Workdir, opts.NotifyFault)
		}
	}

	if serverRequired && len(cfg.Common.Servers) == 0 {
		return nil, NewConfigError("no one nsq server defined. At least one server is required")
	}

	return &cfg, nil
}

// ConfigError represents an error encountered while parsing the configuration file.
type ConfigError struct {
	message string
}

// NewConfigError returns a new instance of configuration error.
func NewConfigError(a ...interface{}) error {
	return &ConfigError{fmt.Sprintln(a...)}
}

// Error returns a string representation of error.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s", e.message)
}
