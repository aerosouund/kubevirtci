package kind

type ContainerClient interface {
	ImagePull(image string) error
	Create(image string) (string, error)
	Run(containerID string) error
	Remove(containerID string) error
}

type Kind interface{}
