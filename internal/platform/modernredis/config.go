package modernredis

import (
	"emperror.dev/errors"
)

// Config holds information necessary for connecting to Redis.
type Config struct {
	Address string

	Password string

	Db int

	Enabled bool

	SentinelEnabled bool

	MasterName string

	MasterPassword string

	SentinelAddress []string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if !c.SentinelEnabled {
		if c.Address == "" {
			return errors.New("modern redis address is required")
		}
	} else {
		if len(c.SentinelAddress) == 0 {
			return errors.New("modern redis sentinel address are required")
		}
	}

	return nil
}
