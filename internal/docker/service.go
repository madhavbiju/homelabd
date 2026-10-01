package docker

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
)

type Service struct {
	cli *client.Client
}

func (s *Service) Client() *client.Client {
	return s.cli
}

func NewService(socketPath string) (*Service, error) {
	opts := []client.Opt{
		client.WithHost("unix://" + socketPath),
		client.WithAPIVersionNegotiation(),
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	return &Service{cli: cli}, nil
}

func (s *Service) Info(ctx context.Context) (system.Info, error) {
	return s.cli.Info(ctx)
}

type ContainerSummary struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Created int64             `json:"created"`
	Labels  map[string]string `json:"labels"`
}

func (s *Service) ListContainers(ctx context.Context) ([]ContainerSummary, error) {
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}

	var results []ContainerSummary
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		results = append(results, ContainerSummary{
			ID:      c.ID,
			Name:    name,
			Image:   c.Image,
			State:   c.State,
			Status:  c.Status,
			Created: c.Created,
			Labels:  c.Labels,
		})
	}
	return results, nil
}

func (s *Service) GetContainer(ctx context.Context, id string) (container.InspectResponse, error) {
	return s.cli.ContainerInspect(ctx, id)
}

func (s *Service) GetLogs(ctx context.Context, id string, tail string) (io.ReadCloser, error) {
	opts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Tail:       tail,
	}
	return s.cli.ContainerLogs(ctx, id, opts)
}

func (s *Service) OperateContainer(ctx context.Context, id, operation string) error {
	switch operation {
	case "start":
		return s.cli.ContainerStart(ctx, id, container.StartOptions{})
	case "stop":
		return s.cli.ContainerStop(ctx, id, container.StopOptions{})
	case "restart":
		return s.cli.ContainerRestart(ctx, id, container.StopOptions{})
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}

func (s *Service) GetStats(ctx context.Context, id string) (io.ReadCloser, error) {
	stats, err := s.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return nil, err
	}
	return stats.Body, nil
}

type ImageSummary struct {
	ID          string   `json:"id"`
	RepoTags    []string `json:"repo_tags"`
	Size        int64    `json:"size"`
	Created     int64    `json:"created"`
}

func (s *Service) ListImages(ctx context.Context) ([]ImageSummary, error) {
	images, err := s.cli.ImageList(ctx, image.ListOptions{All: false})
	if err != nil {
		return nil, err
	}

	var results []ImageSummary
	for _, img := range images {
		results = append(results, ImageSummary{
			ID:       img.ID,
			RepoTags: img.RepoTags,
			Size:     img.Size,
			Created:  img.Created,
		})
	}
	return results, nil
}

func (s *Service) GetImage(ctx context.Context, id string) (image.InspectResponse, error) {
	img, _, err := s.cli.ImageInspectWithRaw(ctx, id)
	return img, err
}

func (s *Service) PullImage(ctx context.Context, ref string) (io.ReadCloser, error) {
	return s.cli.ImagePull(ctx, ref, image.PullOptions{})
}
