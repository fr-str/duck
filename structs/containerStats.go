package structs

type Stats struct {
	// Read       string     `json:"read"`
	// Preread    string     `json:"preread"`
	// PidsStats  PidsStats  `json:"pids_stats"`
	BlkioStats BlkioStats `json:"blkio_stats"`
	// NumProcs     int64        `json:"num_procs"`
	// StorageStats StorageStats `json:"storage_stats"`
	CPUStats CPUStats `json:"cpu_stats"`
	// PrecpuStats CPUStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
	Name        string      `json:"name"`
	ID          string      `json:"id"`
	Networks    Networks    `json:"networks"`
}

type MiniStats struct {
	CPUUsage  float64 // %
	Memory    MiniMem
	Network   MiniNet
	BlockIO   MiniBlk
	Monitored bool `json:"-"`
}

type MiniMem struct {
	Usage float64 // MiB
	Limit float64 // MiB
}

type MiniNet struct {
	I float64 // MiB
	O float64 // MiB
}

type MiniBlk struct {
	I float64 // MiB
	O float64 // MiB
}

type BlkioStats struct {
	IoServiceBytesRecursive []IoServiceBytesRecursive `json:"io_service_bytes_recursive"`
	IoServicedRecursive     interface{}               `json:"io_serviced_recursive"`
	IoQueueRecursive        interface{}               `json:"io_queue_recursive"`
	IoServiceTimeRecursive  interface{}               `json:"io_service_time_recursive"`
	IoWaitTimeRecursive     interface{}               `json:"io_wait_time_recursive"`
	IoMergedRecursive       interface{}               `json:"io_merged_recursive"`
	IoTimeRecursive         interface{}               `json:"io_time_recursive"`
	SectorsRecursive        interface{}               `json:"sectors_recursive"`
}

type IoServiceBytesRecursive struct {
	Major int64  `json:"major"`
	Minor int64  `json:"minor"`
	Op    string `json:"op"`
	Value int64  `json:"value"`
}

type CPUStats struct {
	CPUUsage       CPUUsage       `json:"cpu_usage"`
	SystemCPUUsage float64        `json:"system_cpu_usage"`
	OnlineCpus     uint64         `json:"online_cpus"`
	ThrottlingData ThrottlingData `json:"throttling_data"`
}

type CPUUsage struct {
	TotalUsage        uint64 `json:"total_usage"`
	UsageInKernelmode uint64 `json:"usage_in_kernelmode"`
	UsageInUsermode   uint64 `json:"usage_in_usermode"`
}

type ThrottlingData struct {
	Periods          uint64 `json:"periods"`
	ThrottledPeriods uint64 `json:"throttled_periods"`
	ThrottledTime    uint64 `json:"throttled_time"`
}

type MemoryStats struct {
	Usage uint64            `json:"usage"`
	Stats map[string]uint64 `json:"stats"`
	Limit uint64            `json:"limit"`
}

type Networks struct {
	Eth0 Eth0 `json:"eth0"`
}

type Eth0 struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxPackets uint64 `json:"rx_packets"`
	RxErrors  uint64 `json:"rx_errors"`
	RxDropped uint64 `json:"rx_dropped"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxPackets uint64 `json:"tx_packets"`
	TxErrors  uint64 `json:"tx_errors"`
	TxDropped uint64 `json:"tx_dropped"`
}

type PidsStats struct {
	Current uint64 `json:"current"`
	Limit   uint64 `json:"limit"`
}
