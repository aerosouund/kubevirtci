package bootc

import (
	"embed"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"

	"kubevirt.io/kubevirtci/cluster-provision/gocli/cri"
)

//go:embed k8s-container/k8s.Containerfile
var k8sContainerfile []byte

//go:embed k8s-container/linux.Containerfile
var linuxContainerfile []byte

//go:embed k8s-container/provision-system.sh
var provisionSystem []byte

//go:embed k8s-container/provision-system.service
var provisionSystemService []byte

//go:embed k8s-container/patches
var patches embed.FS

type BootcProvisioner struct {
	cri cri.ContainerClient
}

func NewBootcProvisioner(cri cri.ContainerClient) *BootcProvisioner {
	return &BootcProvisioner{
		cri: cri,
	}
}

func (b *BootcProvisioner) BuildLinuxBase(tag string) error {
	containerFile, err := os.Create("linux.Containerfile")
	if err != nil {
		return err
	}
	_, err = containerFile.Write(linuxContainerfile)
	if err != nil {
		return err
	}

	err = b.cri.Build(tag, "linux.Containerfile")
	if err != nil {
		return err
	}
	return nil
}

func (b *BootcProvisioner) BuildK8sBase(tag, k8sVersion string) error {
	if err := os.Mkdir("patches", 0777); err != nil {
		return err
	}

	err := fs.WalkDir(patches, "patches", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".yaml" {
			yamlData, err := patches.ReadFile(path)
			if err != nil {
				return err
			}

			yamlFile, err := os.Create(path)
			if err != nil {
				return err
			}

			_, err = yamlFile.Write(yamlData)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	containerFile, err := os.Create("k8s.Containerfile")
	if err != nil {
		return err
	}
	_, err = containerFile.Write(k8sContainerfile)
	if err != nil {
		return err
	}

	err = b.cri.Build(tag, "k8s.Containerfile")
	if err != nil {
		return err
	}
	return nil
}

func (b *BootcProvisioner) GenerateQcow(image string) error {
	err := os.Mkdir("output", 0777)
	if err != nil {
		return err
	}

	runArgs := []string{"--rm", "-it",
		"--privileged",
		"--security-opt label=type:unconfined_t",
		"-v output:/output",
		"-v /var/lib/containers/storage:/var/lib/containers/storage",
		"-v config.toml:/config.toml:ro",
		"quay.io/centos-bootc/bootc-image-builder:latest",
		"--type qcow2",
		"--local",
		image}

	err = b.cri.Run(runArgs)
	if err != nil {
		return err
	}

	return nil
}
