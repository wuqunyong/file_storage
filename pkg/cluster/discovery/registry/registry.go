// Package registry is an interface for service discovery
package registry

import (
	"errors"
)

var (

	// Not found error when GetService is called.
	ErrNotFound = errors.New("service not found")
	// Watcher stopped error when watcher is stopped.
	ErrWatcherStopped = errors.New("watcher stopped")
)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}.
type Registry interface {
	Init(...Option) error
	Options() Options
	Register(*Service, ...RegisterOption) error
	Deregister(*Service, ...DeregisterOption) error
	GetService(string, ...GetOption) ([]*Service, error)
	ListServices(...ListOption) ([]*Service, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []*Endpoint       `json:"endpoints"`
	Nodes     []*Node           `json:"nodes"`
}

type Node struct {
	Metadata map[string]string `json:"metadata"`
	Id       string            `json:"id"`
	Address  string            `json:"address"`
}

type Endpoint struct {
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
	Name     string            `json:"name"`
}

type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)

// Register a service node. Additionally supply options such as TTL.
func Register(registry Registry, s *Service, opts ...RegisterOption) error {
	return registry.Register(s, opts...)
}

// Deregister a service node.
func Deregister(registry Registry, s *Service) error {
	return registry.Deregister(s)
}

// Retrieve a service. A slice is returned since we separate Name/Version.
func GetService(registry Registry, name string) ([]*Service, error) {
	return registry.GetService(name)
}

// List the services. Only returns service names.
func ListServices(registry Registry) ([]*Service, error) {
	return registry.ListServices()
}

// Watch returns a watcher which allows you to track updates to the registry.
func Watch(registry Registry, opts ...WatchOption) (Watcher, error) {
	return registry.Watch(opts...)
}

func String(registry Registry) string {
	return registry.String()
}
