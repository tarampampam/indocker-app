package docker_info

type (
	// Container contains response of Engine API.
	// GET "/containers/json"
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/types.go#L152
	Container struct {
		ID         string            `json:"Id"`
		Names      []string          `json:"Names"`
		Image      string            `json:"Image"`
		ImageID    string            `json:"ImageID"`
		Command    string            `json:"Command"`
		Created    int64             `json:"Created"`
		Ports      []Port            `json:"Ports,omitempty"`
		SizeRw     int64             `json:"SizeRw,omitempty"`
		SizeRootFs int64             `json:"SizeRootFs,omitempty"`
		Labels     map[string]string `json:"Labels,omitempty"`
		State      string            `json:"State"`
		Status     string            `json:"Status"`
		HostConfig struct {
			NetworkMode string `json:"NetworkMode,omitempty"`
		} `json:"HostConfig"`
		NetworkSettings *SummaryNetworkSettings `json:"NetworkSettings,omitempty"`
		Mounts          []MountPoint            `json:"Mounts,omitempty"`
	}

	// Port An open port on a container.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/port.go#L8
	Port struct {
		// Host IP address that the container's port is mapped to
		IP string `json:"IP,omitempty"`

		// Port on the container
		// Required: true
		PrivatePort uint16 `json:"PrivatePort"`

		// Port exposed on the host
		PublicPort uint16 `json:"PublicPort,omitempty"`

		// type
		// Required: true
		Type string `json:"Type"`
	}

	// SummaryNetworkSettings provides a summary of container's networks.
	// in /containers/json
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/types.go#L491
	SummaryNetworkSettings struct {
		Networks map[string]*EndpointSettings `json:"Networks"`
	}

	// EndpointSettings stores the network endpoint details.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/network/network.go#L49
	EndpointSettings struct {
		// Configurations
		IPAMConfig *EndpointIPAMConfig `json:"IPAMConfig,omitempty"`
		Links      []string            `json:"Links,omitempty"`
		Aliases    []string            `json:"Aliases,omitempty"`
		// Operational data
		NetworkID           string            `json:"NetworkID"`
		EndpointID          string            `json:"EndpointID"`
		Gateway             string            `json:"Gateway"`
		IPAddress           string            `json:"IPAddress"`
		IPPrefixLen         int               `json:"IPPrefixLen"`
		IPv6Gateway         string            `json:"IPv6Gateway"`
		GlobalIPv6Address   string            `json:"GlobalIPv6Address"`
		GlobalIPv6PrefixLen int               `json:"GlobalIPv6PrefixLen"`
		MacAddress          string            `json:"MacAddress"`
		DriverOpts          map[string]string `json:"DriverOpts,omitempty"`
	}

	// EndpointIPAMConfig represents IPAM configurations for the endpoint.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/network/network.go#L28
	EndpointIPAMConfig struct {
		IPv4Address  string   `json:"IPv4Address,omitempty"`
		IPv6Address  string   `json:"IPv6Address,omitempty"`
		LinkLocalIPs []string `json:"LinkLocalIPs,omitempty"`
	}

	// MountType represents the type of mount.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/mount/mount.go#L8
	MountType string

	// MountPropagation represents the propagation of a mount.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/mount/mount.go#L42
	MountPropagation string

	// MountPoint represents a mount point configuration inside the container.
	// This is used for reporting the mountpoints in use by a container.
	// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/types.go#L524
	MountPoint struct {
		// Type is the type of mount, see `Type<foo>` definitions in
		// github.com/docker/docker/api/types/mount.Type
		Type MountType `json:"Type,omitempty"`

		// Name is the name reference to the underlying data defined by `Source`
		// e.g., the volume name.
		Name string `json:"Name,omitempty"`

		// Source is the source location of the mount.
		//
		// For volumes, this contains the storage location of the volume (within
		// `/var/lib/docker/volumes/`). For bind-mounts, and `npipe`, this contains
		// the source (host) part of the bind-mount. For `tmpfs` mount points, this
		// field is empty.
		Source string `json:"Source"`

		// Destination is the path relative to the container root (`/`) where the
		// Source is mounted inside the container.
		Destination string `json:"Destination"`

		// Driver is the volume driver used to create the volume (if it is a volume).
		Driver string `json:"Driver,omitempty"`

		// Mode is a comma separated list of options supplied by the user when
		// creating the bind/volume mount.
		//
		// The default is platform-specific (`"z"` on Linux, empty on Windows).
		Mode string `json:"Mode"`

		// RW indicates whether the mount is mounted writable (read-write).
		RW bool `json:"RW"`

		// Propagation describes how mounts are propagated from the host into the
		// mount point, and vice-versa. Refer to the Linux kernel documentation
		// for details:
		// https://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt
		//
		// This field is not used on Windows.
		Propagation MountPropagation `json:"Propagation,omitempty"`
	}
)

// Type constants.
// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/mount/mount.go#L11
const (
	// TypeBind is the type for mounting host dir
	TypeBind MountType = "bind"
	// TypeVolume is the type for remote storage volumes
	TypeVolume MountType = "volume"
	// TypeTmpfs is the type for mounting tmpfs
	TypeTmpfs MountType = "tmpfs"
	// TypeNamedPipe is the type for mounting Windows named pipes
	TypeNamedPipe MountType = "npipe"
	// TypeCluster is the type for Swarm Cluster Volumes.
	TypeCluster MountType = "cluster"
)

// https://github.com/moby/moby/blob/b5568723cee5e75060aa265ea8232907f8fe8533/api/types/mount/mount.go#L44
const (
	// PropagationRPrivate RPRIVATE
	PropagationRPrivate MountPropagation = "rprivate"
	// PropagationPrivate PRIVATE
	PropagationPrivate MountPropagation = "private"
	// PropagationRShared RSHARED
	PropagationRShared MountPropagation = "rshared"
	// PropagationShared SHARED
	PropagationShared MountPropagation = "shared"
	// PropagationRSlave RSLAVE
	PropagationRSlave MountPropagation = "rslave"
	// PropagationSlave SLAVE
	PropagationSlave MountPropagation = "slave"
)
