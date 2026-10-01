package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

type Info struct {
	Hostname           string `json:"hostname"`
	OS                 string `json:"os"`
	Architecture       string `json:"architecture"`
	Kernel             string `json:"kernel"`
	Uptime             uint64 `json:"uptime"`
	CPUModel           string `json:"cpu_model"`
	CPUCores           int    `json:"cpu_cores"`
	TotalMemory        uint64 `json:"total_memory"`
	DockerAvailability bool   `json:"docker_availability"`
}

func (s *Service) GetInfo(ctx context.Context) (*Info, error) {
	hInfo, err := host.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	cpuInfo, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get cpu info: %w", err)
	}
	
	cpuModel := "Unknown"
	if len(cpuInfo) > 0 {
		cpuModel = cpuInfo[0].ModelName
	}
	cores, _ := cpu.CountsWithContext(ctx, true)

	vMem, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get mem info: %w", err)
	}

	// Check Docker socket or CLI availability
	dockerAvail := false
	if err := exec.CommandContext(ctx, "docker", "info").Run(); err == nil {
		dockerAvail = true
	}

	return &Info{
		Hostname:           hInfo.Hostname,
		OS:                 hInfo.OS,
		Architecture:       hInfo.KernelArch,
		Kernel:             hInfo.KernelVersion,
		Uptime:             hInfo.Uptime,
		CPUModel:           cpuModel,
		CPUCores:           cores,
		TotalMemory:        vMem.Total,
		DockerAvailability: dockerAvail,
	}, nil
}

type Resources struct {
	CPUUsage    float64 `json:"cpu_usage_percent"`
	LoadAverage struct {
		Load1  float64 `json:"load1"`
		Load5  float64 `json:"load5"`
		Load15 float64 `json:"load15"`
	} `json:"load_average"`
	Memory struct {
		Total       uint64  `json:"total"`
		Used        uint64  `json:"used"`
		Available   uint64  `json:"available"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"memory"`
	Swap struct {
		Total       uint64  `json:"total"`
		Used        uint64  `json:"used"`
		Free        uint64  `json:"free"`
		UsedPercent float64 `json:"used_percent"`
	} `json:"swap"`
	ProcessCount int `json:"process_count"`
}

func (s *Service) GetResources(ctx context.Context) (*Resources, error) {
	var res Resources

	// CPU Usage
	cpuPercents, err := cpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuPercents) > 0 {
		res.CPUUsage = cpuPercents[0]
	}

	// Load Average
	avg, err := load.AvgWithContext(ctx)
	if err == nil {
		res.LoadAverage.Load1 = avg.Load1
		res.LoadAverage.Load5 = avg.Load5
		res.LoadAverage.Load15 = avg.Load15
	}

	// Memory
	vMem, err := mem.VirtualMemoryWithContext(ctx)
	if err == nil {
		res.Memory.Total = vMem.Total
		res.Memory.Used = vMem.Used
		res.Memory.Available = vMem.Available
		res.Memory.UsedPercent = vMem.UsedPercent
	}

	// Swap
	sMem, err := mem.SwapMemoryWithContext(ctx)
	if err == nil {
		res.Swap.Total = sMem.Total
		res.Swap.Used = sMem.Used
		res.Swap.Free = sMem.Free
		res.Swap.UsedPercent = sMem.UsedPercent
	}

	// Process Count
	pids, err := process.PidsWithContext(ctx)
	if err == nil {
		res.ProcessCount = len(pids)
	}

	return &res, nil
}

type Storage struct {
	Mount       string  `json:"mount"`
	Filesystem  string  `json:"filesystem"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Available   uint64  `json:"available"`
	UsagePercent float64 `json:"usage_percent"`
}

func (s *Service) GetStorage(ctx context.Context) ([]Storage, error) {
	partitions, err := disk.PartitionsWithContext(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get partitions: %w", err)
	}

	var results []Storage
	seen := make(map[string]bool)

	for _, p := range partitions {
		// Filter out snaps, overlays, etc. on linux
		if strings.HasPrefix(p.Mountpoint, "/snap") || 
		   strings.HasPrefix(p.Mountpoint, "/var/lib/docker/overlay") {
			continue
		}
		if seen[p.Mountpoint] {
			continue
		}
		seen[p.Mountpoint] = true

		usage, err := disk.UsageWithContext(ctx, p.Mountpoint)
		if err != nil {
			continue
		}

		results = append(results, Storage{
			Mount:        p.Mountpoint,
			Filesystem:   p.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Available:    usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}

	return results, nil
}

type Network struct {
	Name        string   `json:"name"`
	IPAddresses []string `json:"ip_addresses"`
	MACAddress  string   `json:"mac_address"`
	Flags       []string `json:"flags"`
	BytesRecv   uint64   `json:"bytes_recv"`
	BytesSent   uint64   `json:"bytes_sent"`
}

func (s *Service) GetNetwork(ctx context.Context) ([]Network, error) {
	interfaces, err := net.InterfacesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	ioCounters, err := net.IOCountersWithContext(ctx, true)
	ioMap := make(map[string]net.IOCountersStat)
	for _, io := range ioCounters {
		ioMap[io.Name] = io
	}

	var results []Network
	for _, iface := range interfaces {
		var ips []string
		for _, addr := range iface.Addrs {
			ips = append(ips, addr.Addr)
		}

		netIf := Network{
			Name:        iface.Name,
			IPAddresses: ips,
			MACAddress:  iface.HardwareAddr,
			Flags:       iface.Flags,
		}

		if io, ok := ioMap[iface.Name]; ok {
			netIf.BytesRecv = io.BytesRecv
			netIf.BytesSent = io.BytesSent
		}

		results = append(results, netIf)
	}

	return results, nil
}

type ProcessInfo struct {
	PID           int32   `json:"pid"`
	Name          string  `json:"name"`
	Username      string  `json:"username"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemoryPercent float32 `json:"memory_percent"`
	CreateTime    int64   `json:"create_time"`
	Status        string  `json:"status"`
}

func (s *Service) GetProcesses(ctx context.Context, limit int) ([]ProcessInfo, error) {
	procs, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	var results []ProcessInfo
	for i, p := range procs {
		if limit > 0 && i >= limit {
			break
		}

		name, _ := p.NameWithContext(ctx)
		user, _ := p.UsernameWithContext(ctx)
		cpuP, _ := p.CPUPercentWithContext(ctx)
		memP, _ := p.MemoryPercentWithContext(ctx)
		create, _ := p.CreateTimeWithContext(ctx)
		statusStrs, _ := p.StatusWithContext(ctx)
		status := ""
		if len(statusStrs) > 0 {
			status = statusStrs[0]
		}

		results = append(results, ProcessInfo{
			PID:           p.Pid,
			Name:          name,
			Username:      user,
			CPUPercent:    cpuP,
			MemoryPercent: memP,
			CreateTime:    create,
			Status:        status,
		})
	}

	return results, nil
}
