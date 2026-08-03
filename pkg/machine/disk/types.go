package disk

type Disk struct {
	// MountPoint is the filesystem mountpoint backing Path. It is nil when the
	// mountpoint could not be determined.
	MountPoint *string
	// Path is the configured directory whose disk usage is reported.
	Path      string
	Total     uint64
	Available uint64
}
