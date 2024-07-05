package kind

type ContainerClient interface {
	ImagePull(image string) error
	Create(image string, co CreateOpts) (string, error)
	Run(containerID string) error
	Remove(containerID string) error
}

type CreateOpts struct {
	Privileged    bool
	Ports         map[string]string
	RestartPolicy string
	Network       string
	Command       string
	RM            bool
	Capabilities  []string
}

type Kind interface{}
