package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"upm-backend/internal/models"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerService struct {
	client *client.Client
}

func NewDockerService() (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return &DockerService{
		client: cli,
	}, nil
}

// GetRunningContainers returns all running containers
func (d *DockerService) GetRunningContainers() ([]models.Container, error) {
	ctx := context.Background()
	
	// Get all containers (running and stopped)
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []models.Container
	for _, c := range containers {
		container := d.convertToContainer(c)
		result = append(result, container)
	}

	return result, nil
}

// GetContainerByID returns a specific container by ID
func (d *DockerService) GetContainerByID(containerID string) (*models.Container, error) {
	ctx := context.Background()
	
	// Get container details
	containerInfo, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// Convert to our model
	container := d.convertInspectToContainer(containerInfo)
	return &container, nil
}

// GetContainerStats returns real-time stats for a container
func (d *DockerService) GetContainerStats(containerID string) (interface{}, error) {
	ctx := context.Background()
	
	stats, err := d.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}

	return stats, nil
}

// convertToContainer converts Docker container list item to our model
func (d *DockerService) convertToContainer(c types.Container) models.Container {
	// Parse container name (remove leading slash)
	name := c.Names[0]
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}

	// Convert ports
	var ports []models.PortMapping
	for _, port := range c.Ports {
		ports = append(ports, models.PortMapping{
			IP:          port.IP,
			PrivatePort: int(port.PrivatePort),
			PublicPort:  int(port.PublicPort),
			Type:        port.Type,
		})
	}

	// Parse created time
	created := time.Unix(c.Created, 0)

	return models.Container{
		ID:         c.ID,
		Name:       name,
		Image:      c.Image,
		ImageID:    c.ImageID,
		Status:     c.Status,
		State:      c.State,
		Created:    created,
		Ports:      ports,
		Labels:     c.Labels,
		Command:    c.Command,
		SizeRw:     c.SizeRw,
		SizeRootFs: c.SizeRootFs,
		NetworkMode: "default",
	}
}

// convertInspectToContainer converts Docker container inspect to our model
func (d *DockerService) convertInspectToContainer(info types.ContainerJSON) models.Container {
	// Parse container name (remove leading slash)
	name := info.Name
	if strings.HasPrefix(name, "/") {
		name = name[1:]
	}

	// Convert ports
	var ports []models.PortMapping
	for portStr, bindings := range info.NetworkSettings.Ports {
		for _, binding := range bindings {
			// Parse port number from port string (e.g., "80/tcp" -> 80)
			portParts := strings.Split(string(portStr), "/")
			var privatePort int
			if len(portParts) > 0 {
				fmt.Sscanf(portParts[0], "%d", &privatePort)
			}
			
			var publicPort int
			fmt.Sscanf(binding.HostPort, "%d", &publicPort)
			
			ports = append(ports, models.PortMapping{
				IP:          binding.HostIP,
				PrivatePort: privatePort,
				PublicPort:  publicPort,
				Type:        portParts[1],
			})
		}
	}

	// Convert mounts
	var mounts []models.Mount
	for _, mount := range info.Mounts {
		mounts = append(mounts, models.Mount{
			Type:        string(mount.Type),
			Source:      mount.Source,
			Destination: mount.Destination,
			Mode:        mount.Mode,
			RW:          mount.RW,
			Propagation: string(mount.Propagation),
		})
	}

	// Parse times
	created, _ := time.Parse(time.RFC3339Nano, info.Created)
	var startedAt *time.Time
	var finishedAt *time.Time
	
	if info.State.StartedAt != "" {
		if started, err := time.Parse(time.RFC3339Nano, info.State.StartedAt); err == nil {
			startedAt = &started
		}
	}
	
	if info.State.FinishedAt != "" {
		if finished, err := time.Parse(time.RFC3339Nano, info.State.FinishedAt); err == nil {
			finishedAt = &finished
		}
	}

	return models.Container{
		ID:          info.ID,
		Name:        name,
		Image:       info.Config.Image,
		ImageID:     info.Image,
		Status:      info.State.Status,
		State:       info.State.Status,
		Created:     created,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		Ports:       ports,
		Labels:      info.Config.Labels,
		Command:     strings.Join(info.Config.Cmd, " "),
		SizeRw:      *info.SizeRw,
		SizeRootFs:  *info.SizeRootFs,
		NetworkMode: string(info.HostConfig.NetworkMode),
		Mounts:      mounts,
	}
}

// Close closes the Docker client
func (d *DockerService) Close() error {
	return d.client.Close()
}
