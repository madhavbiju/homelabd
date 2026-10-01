package compose

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Stack struct {
	Name       string `json:"name"`
	WorkingDir string `json:"working_dir"`
	Status     string `json:"status"` // e.g. "running", "exited", "mixed"
	Containers int    `json:"containers"`
}

type Service struct {
	cli *client.Client
}

func NewService(dockerClient *client.Client) *Service {
	return &Service{cli: dockerClient}
}

// ListStacks discovers running or stopped stacks by inspecting Docker container labels.
func (s *Service) ListStacks(ctx context.Context) ([]Stack, error) {
	containers, err := s.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	stacksMap := make(map[string]*Stack)

	for _, c := range containers {
		project, ok := c.Labels["com.docker.compose.project"]
		if !ok || project == "" {
			continue
		}
		
		workingDir := c.Labels["com.docker.compose.project.working_dir"]
		
		if _, exists := stacksMap[project]; !exists {
			stacksMap[project] = &Stack{
				Name:       project,
				WorkingDir: workingDir,
				Status:     c.State,
				Containers: 0,
			}
		}

		stack := stacksMap[project]
		stack.Containers++
		
		// If states differ, mark as mixed. If one is running, at least we know it's partially up.
		if stack.Status != c.State {
			if stack.Status != "mixed" {
				stack.Status = "mixed"
			}
		}
	}

	var results []Stack
	for _, stack := range stacksMap {
		results = append(results, *stack)
	}

	return results, nil
}

func (s *Service) GetStack(ctx context.Context, name string) (*Stack, error) {
	stacks, err := s.ListStacks(ctx)
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks {
		if stack.Name == name {
			return &stack, nil
		}
	}

	return nil, fmt.Errorf("stack not found: %s", name)
}

// runComposeCommand executes docker compose with the given arguments in the specified working directory.
func (s *Service) runComposeCommand(ctx context.Context, workingDir string, args ...string) error {
	if workingDir == "" {
		return fmt.Errorf("working directory is required for compose operations")
	}

	cmdArgs := append([]string{"compose"}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("compose command failed: %w, output: %s", err, strings.TrimSpace(string(output)))
	}

	return nil
}

func (s *Service) Up(ctx context.Context, name string) error {
	stack, err := s.GetStack(ctx, name)
	if err != nil {
		return err
	}
	// Run up -d (detached)
	return s.runComposeCommand(ctx, stack.WorkingDir, "up", "-d")
}

func (s *Service) Down(ctx context.Context, name string) error {
	stack, err := s.GetStack(ctx, name)
	if err != nil {
		return err
	}
	// Never run with -v automatically!
	return s.runComposeCommand(ctx, stack.WorkingDir, "down")
}

func (s *Service) Pull(ctx context.Context, name string) error {
	stack, err := s.GetStack(ctx, name)
	if err != nil {
		return err
	}
	return s.runComposeCommand(ctx, stack.WorkingDir, "pull")
}
