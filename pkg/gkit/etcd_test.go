package gkit

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

func TestEtcd(t *testing.T) {
	client, err := NewEtcdClient([]string{"localhost:2379"})
	if err != nil {
		t.Fatalf("failed to create etcd lock: %v", err)
	}
	defer client.Close()

	worker := func(i int, run bool) {
		id := fmt.Sprintf("worker-%d", i)
		val := fmt.Sprintf("10.0.0.%d", i)

		sd, err := NewEtcdDiscovery(EtcdDiscoveryConfig{
			Client:     client,
			Prefix:     "/services",
			Key:        id,
			Val:        val,
			TTLSeconds: 2,
			Callbacks: DiscoveryCallbacks{
				OnStartedDiscovering: func(services []Service) {
					log.Printf("[%s], onstarted, services: %v", id, services)
				},
				OnStoppedDiscovering: func() {
					log.Printf("[%s], onstoped", id)
				},
				OnServiceChanged: func(services []Service, event DiscoveryEvent) {
					log.Printf("[%s], onchanged, services: %v, event: %v", id, services, event)
				},
			},
		})

		if err != nil {
			log.Fatalf("failed to create service etcdiscovery: %v", err)
		}
		defer sd.Close()

		if !run {
			if err = sd.UnRegister(context.Background()); err != nil {
				log.Fatalf("failed to unregister service [%s]: %v", id, err)
			}
			return
		}

		if err := sd.Register(context.Background()); err != nil {
			log.Fatalf("failed to register service [%s]: %v", id, err)
		}

		if err := sd.Watch(context.Background()); err != nil {
			log.Fatalf("failed to watch service: %v", err)
		}
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 3; i++ {
		id := i
		wg.Add(1)
		go func() { worker(id, true) }()
	}

	go func() {
		time.Sleep(2 * time.Second)
		worker(3, true)
	}()

	// unregister
	go func() {
		time.Sleep(4 * time.Second)
		worker(2, false)
	}()

	wg.Wait()
}
