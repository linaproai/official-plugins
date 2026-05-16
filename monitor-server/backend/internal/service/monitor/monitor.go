// Package monitor implements server monitoring collection, storage, query,
// and cleanup services for the monitor-server source plugin. It owns the
// plugin_monitor_server table and runtime sampling logic instead of depending
// on host-internal server-monitor services.
package monitor

import (
	"context"
	"sync"
	"time"

	netutil "github.com/shirou/gopsutil/v4/net"

	entitymodel "lina-plugin-monitor-server/backend/internal/model/entity"
)

// Storage metadata constants for server-monitor persistence.
const (
	colNodeName  = "node_name"
	colNodeIp    = "node_ip"
	colCreatedAt = "created_at"
	colUpdatedAt = "updated_at"
)

// MonitorData represents all collected server metrics.
type MonitorData struct {
	Server  *ServerInfo    `json:"server"`
	CPU     *CPUInfo       `json:"cpu"`
	Memory  *MemoryInfo    `json:"memory"`
	Disks   []*DiskInfo    `json:"disks"`
	Network *NetworkInfo   `json:"network"`
	GoInfo  *GoRuntimeInfo `json:"goInfo"`
}

// ServerInfo represents server basic information.
type ServerInfo struct {
	Hostname  string `json:"hostname"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
	BootTime  string `json:"bootTime"`
	Uptime    uint64 `json:"uptime"`
	StartTime string `json:"startTime"`
}

// CPUInfo represents CPU metrics.
type CPUInfo struct {
	Cores        int     `json:"cores"`
	ModelName    string  `json:"modelName"`
	UsagePercent float64 `json:"usagePercent"`
}

// MemoryInfo represents memory metrics.
type MemoryInfo struct {
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Available    uint64  `json:"available"`
	UsagePercent float64 `json:"usagePercent"`
}

// DiskInfo represents disk metrics.
type DiskInfo struct {
	Path         string  `json:"path"`
	FsType       string  `json:"fsType"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usagePercent"`
}

// NetworkInfo represents network metrics.
type NetworkInfo struct {
	BytesSent uint64  `json:"bytesSent"`
	BytesRecv uint64  `json:"bytesRecv"`
	SendRate  float64 `json:"sendRate"`
	RecvRate  float64 `json:"recvRate"`
}

// GoRuntimeInfo represents Go runtime metrics.
type GoRuntimeInfo struct {
	Version       string  `json:"version"`
	Goroutines    int     `json:"goroutines"`
	ProcessCPU    float64 `json:"processCpu"`
	ProcessMemory float64 `json:"processMemory"`
	GCPauseNs     uint64  `json:"gcPauseNs"`
	GfVersion     string  `json:"gfVersion"`
	ServiceUptime string  `json:"serviceUptime"`
}

// DBInfo represents database metrics.
type DBInfo struct {
	Version      string `json:"version"`
	MaxOpenConns int    `json:"maxOpenConns"`
	OpenConns    int    `json:"openConns"`
	InUse        int    `json:"inUse"`
	Idle         int    `json:"idle"`
}

// NodeMonitorData wraps monitor data with node info.
type NodeMonitorData struct {
	NodeName  string       `json:"nodeName"`
	NodeIp    string       `json:"nodeIp"`
	Data      *MonitorData `json:"data"`
	CollectAt string       `json:"collectAt"`
}

// serverMonitorRecord reuses the plugin-local generated plugin_monitor_server entity.
type serverMonitorRecord = entitymodel.Server

// Service defines node-local server metric sampling and plugin-owned snapshot storage.
type Service interface {
	// CollectAndStore samples metrics for the current node and upserts the plugin
	// snapshot row. Storage errors are logged; callers do not receive them.
	CollectAndStore(ctx context.Context)
	// Collect gathers all server, CPU, memory, disk, network, and Go runtime
	// metrics from the current process and node without touching persisted state.
	Collect(ctx context.Context) *MonitorData
	// GetDBInfo collects database pool and version metrics on demand. On database
	// errors it returns partial information rather than changing cache or tenant state.
	GetDBInfo(ctx context.Context) *DBInfo
	// GetLatest returns latest plugin snapshot records, optionally for one node.
	// It is a read-only query and returns database or JSON decoding errors.
	GetLatest(ctx context.Context, nodeName string) ([]*NodeMonitorData, error)
	// CleanupStale deletes plugin snapshot records older than threshold and
	// returns the affected count; no i18n, cache, or permission state is modified.
	CleanupStale(ctx context.Context, threshold time.Duration) (int64, error)
}

// Ensure serviceImpl implements Service.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service.
type serviceImpl struct {
	networkMu     sync.Mutex
	startTime     time.Time
	lastNetBytes  *netutil.IOCountersStat
	lastCollectAt time.Time
}

// New creates and returns a new server-monitor service instance.
func New() Service {
	return &serviceImpl{startTime: time.Now()}
}
