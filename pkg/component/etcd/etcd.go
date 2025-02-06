package etcd

import (
	"log/slog"
	"time"

	"github.com/wuqunyong/file_storage/pkg/cluster/discovery/etcd"
	"github.com/wuqunyong/file_storage/pkg/cluster/discovery/registry"
	"github.com/wuqunyong/file_storage/pkg/concepts"
)

type EtcdServiceDiscovery struct {
	registry    registry.Registry
	watcher     registry.Watcher
	engine      concepts.IEngine
	registerTTL time.Duration
}

// etcdctl.exe put /micro/registry/test2/2 {\"name\":\"hello\"}
func NewEtcvServiceDiscovery(opts ...registry.Option) *EtcdServiceDiscovery {
	r := etcd.NewRegistry(opts...)
	return &EtcdServiceDiscovery{
		registry:    r,
		registerTTL: 60 * time.Second,
	}
}

func (sd *EtcdServiceDiscovery) Name() string {
	return "etcd"
}

func (sd *EtcdServiceDiscovery) Priority() int32 {
	return 1
}

func (sd *EtcdServiceDiscovery) SetEngine(engine concepts.IEngine) {
	sd.engine = engine
}

func (sd *EtcdServiceDiscovery) GetEngine() concepts.IEngine {
	return sd.engine
}

func (sd *EtcdServiceDiscovery) OnInit() error {
	service := &registry.Service{
		Name:    "test1",
		Version: "1.0.1",
		Nodes: []*registry.Node{
			{
				Id:      "test1-1",
				Address: "10.0.0.1:10001",
				Metadata: map[string]string{
					"foo": "bar",
				},
			},
		},
	}
	rOpts := []registry.RegisterOption{registry.RegisterTTL(sd.registerTTL)}
	if err := sd.registry.Register(service, rOpts...); err != nil {
		return err
	}

	watcher, err := sd.registry.Watch()
	if err != nil {
		return err
	}
	sd.watcher = watcher

	return nil
}

func (sd *EtcdServiceDiscovery) OnStart() {
	go func() {
		for {
			res, err := sd.watcher.Next()
			if err != nil {
				slog.Error("EtcdServiceDiscovery", "Next", res)
			}

			slog.Info("EtcdServiceDiscovery", "Next", res, "Service", res.Service)
		}
	}()
}

func (sd *EtcdServiceDiscovery) OnCleanup() {

}
