package etcd

import (
	"time"

	"github.com/wuqunyong/file_storage/pkg/cluster/discovery/etcd"
	"github.com/wuqunyong/file_storage/pkg/cluster/discovery/registry"
	"github.com/wuqunyong/file_storage/pkg/concepts"
	"github.com/wuqunyong/file_storage/pkg/logger"
)

var (
	DefaultRegisterInterval    = time.Second * 30
	DefaultRegisterTTL         = time.Second * 90
	DefaultSyncServersInterval = time.Second * 60 * 5
	serviceName                = "test1"
)

type EtcdServiceDiscovery struct {
	registry           registry.Registry
	watcher            registry.Watcher
	engine             concepts.IEngine
	registerTTL        time.Duration
	registerInterval   time.Duration
	service            *registry.Service
	exit               chan chan error
	syncServersRunning chan bool
	running            bool
}

// etcdctl.exe put /micro/registry/test2/2 {\"name\":\"hello\"}
func NewEtcvServiceDiscovery(opts ...registry.Option) *EtcdServiceDiscovery {
	r := etcd.NewRegistry(opts...)
	return &EtcdServiceDiscovery{
		registry:           r,
		registerTTL:        DefaultRegisterTTL,
		registerInterval:   DefaultRegisterInterval,
		exit:               make(chan chan error),
		syncServersRunning: make(chan bool),
		service: &registry.Service{
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
		},
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
	if err := sd.Register(); err != nil {
		return err
	}

	watcher, err := sd.registry.Watch()
	if err != nil {
		return err
	}
	sd.watcher = watcher
	go sd.watch()

	err = sd.SyncServers()
	if err != nil {
		return err
	}
	return nil
}

func (sd *EtcdServiceDiscovery) Register() error {
	rOpts := []registry.RegisterOption{registry.RegisterTTL(sd.registerTTL)}
	if err := sd.registry.Register(sd.service, rOpts...); err != nil {
		return err
	}

	return nil
}

func (sd *EtcdServiceDiscovery) registrar() {
	// Only process if it exists
	ticker := new(time.Ticker)
	if sd.registerInterval > time.Duration(0) {
		ticker = time.NewTicker(sd.registerInterval)
	}

	for {
		bExit := false
		select {
		// Register self on interval
		case <-ticker.C:
			if err := sd.Register(); err != nil {
				logger.Log(logger.ErrorLevel, "EtcdServiceDiscovery", "err", err)
			}
		case ch := <-sd.exit:
			err := sd.registry.Deregister(sd.service)
			ch <- err
			bExit = true
		}

		if bExit {
			break
		}
	}
}

func (sd *EtcdServiceDiscovery) watch() {
	syncServersState := <-sd.syncServersRunning
	for syncServersState {
		syncServersState = <-sd.syncServersRunning
	}

	for {
		res, err := sd.watcher.Next()
		if err != nil {
			logger.Log(logger.ErrorLevel, "EtcdServiceDiscovery", "Next", res)
		}

		logger.Log(logger.InfoLevel, "EtcdServiceDiscovery", "Next", res, "Service", res.Service)
	}
}

func (sd *EtcdServiceDiscovery) SyncServers() error {
	sd.syncServersRunning <- true
	defer func() {
		sd.syncServersRunning <- false
	}()
	start := time.Now()

	s, err := sd.registry.GetService(serviceName)
	if err != nil {
		return err
	}
	for _, service := range s {
		logger.Log(logger.InfoLevel, "SyncServers", "value", service)
	}

	elapsed := time.Since(start)
	logger.Log(logger.InfoLevel, "SyncServers took", "elapsed", elapsed)
	return nil
}

func (sd *EtcdServiceDiscovery) OnStart() {
	if sd.running {
		return
	}
	sd.running = true

	go sd.registrar()
	go sd.watch()
}

func (sd *EtcdServiceDiscovery) OnCleanup() {
	if !sd.running {
		return
	}
	sd.running = false

	// exit and return err
	ch := make(chan error)
	sd.exit <- ch
	err := <-ch
	logger.Log(logger.ErrorLevel, "EtcdServiceDiscovery OnCleanup", "err", err)
}
