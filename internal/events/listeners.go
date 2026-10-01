package events

import (
	"context"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
	"github.com/madhavbiju/homelabd/internal/system"
)

// StartDockerListener listens to the Docker event stream and publishes them to the broker.
func StartDockerListener(ctx context.Context, cli *client.Client, broker *Broker) {
	if cli == nil {
		return
	}

	msgs, errs := cli.Events(ctx, events.ListOptions{})

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				if err != nil {
					slog.Error("Docker event stream error", "error", err)
					// In a production scenario, we might want to backoff and reconnect.
					// For now, sleep briefly and return (it's fatal to the listener).
					time.Sleep(5 * time.Second)
					return 
				}
			case msg := <-msgs:
				// Map docker events to homelabd events
				switch msg.Type {
				case "container":
					// e.g. container.started, container.died
					eventType := "container." + string(msg.Action)
					broker.Publish(eventType, map[string]interface{}{
						"id":     msg.Actor.ID,
						"image":  msg.Actor.Attributes["image"],
						"name":   msg.Actor.Attributes["name"],
						"status": msg.Status,
					})
				case "image":
					eventType := "image." + string(msg.Action)
					broker.Publish(eventType, map[string]interface{}{
						"id":   msg.Actor.ID,
						"name": msg.Actor.Attributes["name"],
					})
				}
			}
		}
	}()
}

// StartSystemPoller periodically polls system resources and publishes changes.
func StartSystemPoller(ctx context.Context, sysSvc *system.Service, broker *Broker) {
	ticker := time.NewTicker(5 * time.Second)
	
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				res, err := sysSvc.GetResources(ctx)
				if err != nil {
					slog.Error("Failed to poll system resources for events", "error", err)
					continue
				}
				broker.Publish("system.resource_changed", res)
			}
		}
	}()
}
