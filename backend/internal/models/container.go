package models

import (
	"time"
)

type Container struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	ImageID     string            `json:"image_id"`
	Status      string            `json:"status"`
	State       string            `json:"state"`
	Created     time.Time         `json:"created"`
	StartedAt   *time.Time        `json:"started_at,omitempty"`
	FinishedAt  *time.Time        `json:"finished_at,omitempty"`
	Ports       []PortMapping     `json:"ports"`
	Labels      map[string]string `json:"labels"`
	Command     string            `json:"command"`
	SizeRw      int64             `json:"size_rw"`
	SizeRootFs  int64             `json:"size_root_fs"`
	NetworkMode string            `json:"network_mode"`
	Mounts      []Mount           `json:"mounts"`
}

type PortMapping struct {
	IP          string `json:"ip"`
	PrivatePort int    `json:"private_port"`
	PublicPort  int    `json:"public_port"`
	Type        string `json:"type"`
}

type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
	RW          bool   `json:"rw"`
	Propagation string `json:"propagation"`
}

type ContainerListResponse struct {
	Containers []Container `json:"containers"`
	Count      int         `json:"count"`
}
