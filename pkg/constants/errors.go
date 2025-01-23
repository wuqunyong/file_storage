package constants

import "errors"

// Errors that can occur during message handling.
var (
	ErrBindingNotFound                = errors.New("binding for this user was not found in etcd")
	ErrEtcdGrantLeaseTimeout          = errors.New("timed out waiting for etcd lease grant")
	ErrEtcdLeaseNotFound              = errors.New("etcd lease not found in group")
	ErrIllegalUID                     = errors.New("illegal uid")
	ErrNoBindingStorageModule         = errors.New("for sending remote pushes or using unique session module while using grpc you need to pass it a BindingStorage")
	ErrNoConnectionToServer           = errors.New("rpc client has no connection to the chosen server")
	ErrNoContextFound                 = errors.New("no context found")
	ErrNoNatsConnectionString         = errors.New("you have to provide a nats url")
	ErrNoServerTypeChosenForRPC       = errors.New("no server type chosen for sending RPC, send a full route in the format server.service.component")
	ErrNoServerWithID                 = errors.New("can't find any server with the provided ID")
	ErrNoServersAvailableOfType       = errors.New("no servers available of this type")
	ErrNotImplemented                 = errors.New("method not implemented")
	ErrRPCRequestTimeout              = errors.New("rpc client: request timeout")
	ErrRPCClientNotInitialized        = errors.New("RPC client is not running")
	ErrRPCJobAlreadyRegistered        = errors.New("rpc job was already registered")
	ErrRPCLocal                       = errors.New("RPC must be to a different server type")
	ErrRPCServerNotInitialized        = errors.New("RPC server is not running")
	ErrReplyShouldBeNotNull           = errors.New("reply must not be null")
	ErrReplyShouldBePtr               = errors.New("reply must be a pointer")
	ErrRequestOnNotify                = errors.New("tried to request a notify route")
	ErrRouterNotInitialized           = errors.New("router is not initialized")
	ErrServerNotFound                 = errors.New("server not found")
	ErrServiceDiscoveryNotInitialized = errors.New("service discovery client is not initialized")
	ErrConnectionClosed               = errors.New("client connection closed")
)
