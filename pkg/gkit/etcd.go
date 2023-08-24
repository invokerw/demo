package gkit

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"strings"
	"sync"
	"time"
)

func NewEtcdClient(endpoints []string) (*etcd.Client, error) {
	c, err := etcd.New(etcd.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "NewEtcdClient failed")
	}
	return c, nil
}

func NewEtcdDiscovery(config EtcdDiscoveryConfig) (*EtcdDiscovery, error) {
	session, err := concurrency.NewSession(config.Client, concurrency.WithTTL(config.TTLSeconds))
	if err != nil {
		return nil, err
	}
	config.Prefix = strings.TrimSuffix(config.Prefix, "/") + "/"
	return &EtcdDiscovery{
		EtcdDiscoveryConfig: config,
		session:             session,
		realKey:             config.Prefix + config.Key,
		services:            make(map[string]string),
	}, nil
}

type EtcdDiscoveryConfig struct {
	Client     *etcd.Client
	Prefix     string
	Key        string
	Val        string
	TTLSeconds int

	Callbacks DiscoveryCallbacks
}

type DiscoveryCallbacks struct {
	OnStartedDiscovering func(services []Service)
	OnServiceChanged     func(services []Service, event DiscoveryEvent)
	OnStoppedDiscovering func()
}

type EtcdDiscovery struct {
	EtcdDiscoveryConfig
	realKey string

	session *concurrency.Session

	watchContext context.Context
	watchCancel  context.CancelFunc

	services map[string]string
	mu       sync.RWMutex
}

type DiscoveryEvent struct {
	Type mvccpb.Event_EventType
	Service
}

type Service struct {
	Path string
	Name string
	Val  string
}

func (d *EtcdDiscovery) Register(ctx context.Context) error {
	lease := d.session.Lease()

	_, err := d.Client.Put(ctx, d.realKey, d.Val, etcd.WithLease(lease))

	return err
}

func (d *EtcdDiscovery) UnRegister(ctx context.Context) error {
	_, err := d.Client.Delete(ctx, d.realKey)
	return err
}

func (d *EtcdDiscovery) Close() error {
	if d.watchCancel != nil {
		d.watchCancel()
	}
	return d.session.Close()
}

func (d *EtcdDiscovery) Watch(ctx context.Context) error {
	d.watchContext, d.watchCancel = context.WithCancel(ctx)

	resp, err := d.Client.Get(d.watchContext, d.Prefix, etcd.WithPrefix())
	if err != nil {
		return err
	}

	services := make(map[string]string)
	for _, kv := range resp.Kvs {
		services[string(kv.Key)] = string(kv.Value)
	}
	d.setServices(services)

	if d.Callbacks.OnStartedDiscovering != nil {
		d.Callbacks.OnStartedDiscovering(d.ListServices())
	}

	defer func() {
		if d.Callbacks.OnStoppedDiscovering != nil {
			d.Callbacks.OnStoppedDiscovering()
		}
	}()

	defer d.watchCancel()

	ch := d.Client.Watch(d.watchContext, d.Prefix, etcd.WithPrefix())
	for {
		select {
		case <-d.watchContext.Done():
			return nil
		case wr, ok := <-ch:
			if !ok {
				return fmt.Errorf("watch closed")
			}
			if wr.Err() != nil {
				return wr.Err()
			}
			for _, ev := range wr.Events {
				key, val := string(ev.Kv.Key), string(ev.Kv.Value)
				switch ev.Type {
				case mvccpb.PUT:
					d.addService(key, val)
				case mvccpb.DELETE:
					d.delService(key)
				}
				if d.Callbacks.OnServiceChanged != nil {
					event := DiscoveryEvent{Type: ev.Type, Service: d.serviceFromKv(key, val)}
					d.Callbacks.OnServiceChanged(d.ListServices(), event)
				}
			}
		}
	}
}

func (d *EtcdDiscovery) serviceFromKv(key, val string) Service {
	return Service{
		Path: key,
		Name: strings.TrimPrefix(key, d.Prefix),
		Val:  val,
	}
}

func (d *EtcdDiscovery) ListServices() []Service {
	d.mu.RLock()
	defer d.mu.RUnlock()

	items := make([]Service, 0, len(d.services))
	for k, v := range d.services {
		items = append(items, d.serviceFromKv(k, v))
	}

	return items
}

func (d *EtcdDiscovery) DrainServices(ctx context.Context) error {
	_, err := d.Client.Delete(ctx, d.Prefix, etcd.WithPrefix())
	return err
}

func (d *EtcdDiscovery) setServices(services map[string]string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.services = services
}

func (d *EtcdDiscovery) addService(key, val string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.services[key] = val
}

func (d *EtcdDiscovery) delService(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.services, key)
}
