package zkdiscovery

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/qingwave/gocorex/x/containerx"
)

const defaultTimeout = 5 * time.Second

func New(config ZkDiscoveryConfig) (*ZkDiscovery, error) {
	if len(config.ACL) == 0 {
		config.ACL = zk.WorldACL(zk.PermAll)
	}

	if config.SessionTimeout == 0 {
		config.SessionTimeout = defaultTimeout
	}

	conn, _, err := zk.Connect(config.Endpoints, config.SessionTimeout)
	if err != nil {
		return nil, err
	}

	config.Path = strings.TrimSuffix(config.Path, "/")

	if err := createZkPath(conn, config.Path, config.ACL); err != nil {
		return nil, err
	}

	return &ZkDiscovery{
		ZkDiscoveryConfig: config,
		conn:              conn,
		myKey:             config.Path + "/" + config.Key,
		services:          containerx.Set[string]{},
	}, nil
}

type ZkDiscoveryConfig struct {
	Endpoints      []string
	SessionTimeout time.Duration
	ACL            []zk.ACL
	Path           string
	Key            string
	Val            string

	watchContext context.Context
	watchCancel  context.CancelFunc

	Callbacks DiscoveryCallbacks
}

type DiscoveryCallbacks struct {
	OnStartedDiscovering func(services []Service)
	OnServiceChanged     func(services []Service)
	OnStoppedDiscovering func()
}

type ZkDiscovery struct {
	ZkDiscoveryConfig
	myKey string

	conn *zk.Conn

	services containerx.Set[string]
	mu       sync.RWMutex
}

type Service struct {
	Name string
	Val  string
}

func (d *ZkDiscovery) Register(ctx context.Context) error {
	_, err := d.conn.Create(d.myKey, []byte(d.Val), 1, d.ACL)
	if err == zk.ErrNodeExists {
		return nil
	}
	return err
}

func (d *ZkDiscovery) UnRegister(ctx context.Context) error {
	err := d.conn.Delete(d.myKey, -1)
	if err == zk.ErrNoNode {
		return nil
	}
	return err
}

func (d *ZkDiscovery) Close() error {
	if d.watchCancel != nil {
		d.watchCancel()
	}

	d.conn.Close()
	return nil
}

func (d *ZkDiscovery) Watch(ctx context.Context) error {
	d.watchContext, d.watchCancel = context.WithCancel(ctx)

	if err := d.refreshServices(); err != nil {
		return err
	}

	if d.Callbacks.OnStartedDiscovering != nil {
		d.Callbacks.OnStartedDiscovering(d.ListServices())
	}

	defer d.watchCancel()

	defer func() {
		if d.Callbacks.OnStoppedDiscovering != nil {
			d.Callbacks.OnStoppedDiscovering()
		}
	}()

loop:
	_, _, ch, err := d.conn.ChildrenW(d.Path)
	if err != nil {
		return err
	}
	for {
		select {
		case <-d.watchContext.Done():
			return nil
		case e, ok := <-ch:
			if !ok {
				goto loop
			}
			if e.Err != nil {
				return e.Err
			}

			switch e.Type {
			case zk.EventNodeCreated, zk.EventNodeChildrenChanged:
				d.refreshServices()
			}

			switch e.State {
			case zk.StateExpired:
				return fmt.Errorf("node [%s] expired", d.myKey)
			case zk.StateDisconnected:
				return nil
			}

			if d.Callbacks.OnServiceChanged != nil {
				d.Callbacks.OnServiceChanged(d.ListServices())
			}
		}
	}
}

func (d *ZkDiscovery) serviceFromKv(key, val string) Service {
	return Service{
		Name: strings.TrimPrefix(key, d.Path),
		Val:  val,
	}
}

func (d *ZkDiscovery) ListServices() []Service {
	d.mu.RLock()
	defer d.mu.RUnlock()

	items := make([]Service, 0, len(d.services))
	for k := range d.services {
		items = append(items, d.serviceFromKv(k, ""))
	}

	return items
}

func (d *ZkDiscovery) DrainServices(ctx context.Context) error {
	return d.conn.Delete(d.Path, -1)
}

func (d *ZkDiscovery) refreshServices() error {
	children, _, err := d.conn.Children(d.Path)
	if err != nil {
		return err
	}

	services := containerx.NewSet(children...)
	d.setServices(services)
	return nil
}

func (d *ZkDiscovery) setServices(services containerx.Set[string]) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.services = services
}

func (d *ZkDiscovery) addService(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.services.Insert(key)
}

func (d *ZkDiscovery) delService(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.services, key)
}

func createZkPath(conn *zk.Conn, path string, acl []zk.ACL) error {
	if path == "" || path == "/" {
		return nil
	}

	ok, _, _ := conn.Exists(path)
	if ok {
		return nil
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	var node string
	for _, part := range parts {
		node += "/" + part
		_, err := conn.Create(node, []byte{}, 0, acl)
		if err != nil && err != zk.ErrNodeExists {
			return err
		}
	}

	return nil
}
