// Package registry is a component for service discovery
package registry

import (
	"errors"
	"fmt"
	"strings"

	"log/slog"

	"github.com/go-orb/go-orb/cli"
	"github.com/go-orb/go-orb/config"
	"github.com/go-orb/go-orb/log"
	"github.com/go-orb/go-orb/types"
)

// ComponentType is the components name.
const ComponentType = "registry"

var (
	// ErrNotFound is a not found error when GetService is called.
	ErrNotFound = errors.New("service not found")
	// ErrWatcherStopped is a error when watcher is stopped.
	ErrWatcherStopped = errors.New("watcher stopped")
)

// Registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}.
type Registry interface {
	types.Component

	// ServiceName returns types.ServiceName that has been provided to Provide().
	ServiceName() string

	// ServiceVersion returns types.ServiceVersion that has been provided to Provide().
	ServiceVersion() string

	// Register registers a service within the registry.
	Register(srv *Service, opts ...RegisterOption) error

	// Deregister deregisters a service within the registry.
	Deregister(srv *Service, opts ...DeregisterOption) error

	// GetService returns a service from the registry.
	GetService(name string, opts ...GetOption) ([]*Service, error)

	// ListServices lists services within the registry.
	ListServices(opts ...ListOption) ([]*Service, error)

	// Watch returns a Watcher which you can watch on.
	Watch(opts ...WatchOption) (Watcher, error)
}

// Type is the registry type it is returned when you use ProvideRegistry
// which selects a registry to use based on the plugin configuration.
type Type struct {
	Registry
}

// Service represents a service in a registry.
type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []*Endpoint       `json:"endpoints"`
	Nodes     []*Node           `json:"nodes"`
}

// String returns a string representation of the Service.
func (s *Service) String() string {
	nodes := []string{}
	for _, n := range s.Nodes {
		nodes = append(nodes, n.String())
	}

	return fmt.Sprintf("Service{Name: %s, Version: %s, Nodes: (%s), Endpoints: %d}",
		s.Name, s.Version, strings.Join(nodes, ", "), len(s.Endpoints))
}

// Node represents a service node in a registry.
// One service can be comprised of multiple nodes.
type Node struct {
	ID string `json:"id"`
	// ip:port
	Address string `json:"address"`
	// grpc/h2c/http/http3 uvm., since go-orb!
	Transport string            `json:"transport"`
	Metadata  map[string]string `json:"metadata"`
}

// String returns a string representation of the Node.
func (n *Node) String() string {
	return fmt.Sprintf("Node{ID: %s, Address: %s, Transport: %s}",
		n.ID, n.Address, n.Transport)
}

// Endpoint represents a service endpoint in a registry.
type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

// String returns a string representation of the Endpoint.
func (e *Endpoint) String() string {
	var reqName, respName string
	if e.Request != nil {
		reqName = e.Request.Name
	}

	if e.Response != nil {
		respName = e.Response.Name
	}

	return fmt.Sprintf("Endpoint{Name: %s, Request: %s, Response: %s}",
		e.Name, reqName, respName)
}

// Value is a value container used in the registry.
type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

// String returns a string representation of the Value.
func (v *Value) String() string {
	return fmt.Sprintf("Value{Name: %s, Type: %s, Values: %d}",
		v.Name, v.Type, len(v.Values))
}

// New creates a new registry without side-effects.
func New(
	name string,
	version string,
	configData map[string]any,
	components *types.Components,
	logger log.Logger,
	opts ...Option,
) (Type, error) {
	cfg := NewConfig(opts...)

	if err := config.Parse(nil, DefaultConfigSection, configData, &cfg); err != nil && !errors.Is(err, config.ErrNoSuchKey) {
		return Type{}, err
	}

	if cfg.Plugin == "" {
		logger.Warn("empty registry plugin, using the default", "default", DefaultRegistry)
		cfg.Plugin = DefaultRegistry
	}

	logger.Debug("Registry", "plugin", cfg.Plugin)

	provider, ok := Plugins.Get(cfg.Plugin)
	if !ok {
		return Type{}, fmt.Errorf("Registry plugin '%s' not found, did you import it?", cfg.Plugin)
	}

	// Configure the logger.
	cLogger, err := logger.WithConfig([]string{DefaultConfigSection}, configData)
	if err != nil {
		return Type{}, err
	}

	cLogger = cLogger.With(slog.String("component", ComponentType), slog.String("plugin", cfg.Plugin))

	instance, err := provider(name, version, configData, components, cLogger, opts...)
	if err != nil {
		return Type{}, err
	}

	return Type{Registry: instance}, nil
}

// Provide is the registry provider for wire.
// It parses the config from "configs", fetches the "Plugin" from the config and
// then forwards all it's arguments to the factory which it get's from "Plugins".
func Provide(
	svcCtx *cli.ServiceContext,
	components *types.Components,
	logger log.Logger,
	opts ...Option,
) (Type, error) {
	reg, err := New(svcCtx.Name(), svcCtx.Version(), svcCtx.Config, components, logger, opts...)
	if err != nil {
		return Type{}, err
	}

	// Register the registry as a component.
	err = components.Add(&reg, types.PriorityRegistry)
	if err != nil {
		logger.Warn("while registering registry as a component", "error", err)
	}

	return reg, nil
}

// ProvideNoOpts is the registry provider for wire without options.
func ProvideNoOpts(
	svcCtx *cli.ServiceContext,
	components *types.Components,
	logger log.Logger,
) (Type, error) {
	return Provide(svcCtx, components, logger)
}
