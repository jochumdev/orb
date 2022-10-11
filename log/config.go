package log

import (
	"github.com/hashicorp/go-multierror"
	"jochum.dev/jochumdev/orb/config/chelp"
)

const (
	CONFIG_KEY_FIELDS          = "fields"
	CONFIG_KEY_LEVEL           = "level"
	CONFIG_KEY_CALLERSKIPFRAME = "caller_skip_frame"
)

type Config interface {
	chelp.PluginConfig

	Fields() map[string]any
	Level() string
	CallerSkipFrame() int
}

type BaseConfig struct {
	*chelp.BasicPlugin
	fields          map[string]any
	level           string
	callerSkipFrame int
}

func NewBaseConfig() Config {
	return &BaseConfig{
		BasicPlugin: chelp.NewBasicPlugin(),
	}
}

func (c *BaseConfig) Load(m map[string]any) error {
	var result error

	if err := c.BasicPlugin.Load(m); err != nil {
		result = multierror.Append(err)
	}

	// Optionals
	var err error
	if c.fields, err = chelp.Get(m, CONFIG_KEY_FIELDS, map[string]any{}); err != nil && err != chelp.ErrNotExistant {
		result = multierror.Append(err)
	}
	if c.level, err = chelp.Get(m, CONFIG_KEY_LEVEL, ""); err != nil && err != chelp.ErrNotExistant {
		result = multierror.Append(err)
	}
	if c.callerSkipFrame, err = chelp.Get(m, CONFIG_KEY_CALLERSKIPFRAME, 0); err != nil && err != chelp.ErrNotExistant {
		result = multierror.Append(err)
	}

	return result
}

func (c *BaseConfig) Store(m map[string]any) error {
	if err := c.BasicPlugin.Store(m); err != nil {
		return err
	}

	m[CONFIG_KEY_FIELDS] = c.fields
	m[CONFIG_KEY_LEVEL] = c.level
	m[CONFIG_KEY_CALLERSKIPFRAME] = c.callerSkipFrame

	return nil
}

func (c *BaseConfig) Fields() map[string]any { return c.fields }
func (c *BaseConfig) Level() string          { return c.level }
func (c *BaseConfig) CallerSkipFrame() int   { return c.callerSkipFrame }
