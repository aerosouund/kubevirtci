package providers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

func NewKubevirtProvider(k8sversion string, image string, cli *client.Client, options ...KubevirtProviderOption) *KubevirtProvider {
	bp := &KubevirtProvider{
		Version:     k8sversion,
		Nodes:       1,
		Numa:        1,
		Memory:      "3096M",
		CPU:         2,
		Background:  true,
		Image:       image,
		RandomPorts: true,
		Docker:      cli,
	}

	for _, option := range options {
		option(bp)
	}

	return bp
}

func (kp *KubevirtProvider) SetClient() {}

func (kp *KubevirtProvider) Start(ctx context.Context, cancel context.CancelFunc, portMap nat.PortMap) (retErr error) {
	stop := make(chan error, 10)
	containers, _, done := docker.NewCleanupHandler(kp.Docker, stop, os.Stdout, false)

	defer func() {
		stop <- retErr
		<-done
	}()

	go kp.handleInterrupt(cancel, stop)

	dnsmasq, err := kp.RunDNSMasq(ctx, portMap)
	if err != nil {
		return err
	}
	kp.DNSMasq = dnsmasq
	containers <- dnsmasq

	dnsmasqJSON, err := kp.Docker.ContainerInspect(context.Background(), kp.DNSMasq)
	if err != nil {
		return err
	}

	sshPort, err := utils.GetPublicPort(utils.PortSSH, dnsmasqJSON.NetworkSettings.Ports)
	apiServerPort, err := utils.GetPublicPort(utils.PortAPI, dnsmasqJSON.NetworkSettings.Ports)

	registry, err := kp.RunRegistry(ctx)
	if err != nil {
		return err
	}
	containers <- registry

	if kp.NFSData != "" {
		nfsGanesha, err := kp.RunNFSGanesha(ctx)
		if err != nil {
			return nil
		}
		containers <- nfsGanesha
	}

	return nil
}

func (kp *KubevirtProvider) RunDNSMasq(ctx context.Context, portMap nat.PortMap) (string, error) {
	dnsmasqMounts := []mount.Mount{}
	_, err := os.Stat("/lib/modules")
	if err == nil {
		dnsmasqMounts = []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "/lib/modules",
				Target: "/lib/modules",
			},
		}

	}

	dnsmasq, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: kp.Image,
		Env: []string{
			fmt.Sprintf("NUM_NODES=%d", kp.Nodes),
			fmt.Sprintf("NUM_SECONDARY_NICS=%d", kp.SecondaryNics),
		},
		Cmd: []string{"/bin/bash", "-c", "/dnsmasq.sh"},
		ExposedPorts: nat.PortSet{
			utils.TCPPortOrDie(utils.PortSSH):         {},
			utils.TCPPortOrDie(utils.PortRegistry):    {},
			utils.TCPPortOrDie(utils.PortOCP):         {},
			utils.TCPPortOrDie(utils.PortAPI):         {},
			utils.TCPPortOrDie(utils.PortVNC):         {},
			utils.TCPPortOrDie(utils.PortHTTP):        {},
			utils.TCPPortOrDie(utils.PortHTTPS):       {},
			utils.TCPPortOrDie(utils.PortPrometheus):  {},
			utils.TCPPortOrDie(utils.PortGrafana):     {},
			utils.TCPPortOrDie(utils.PortUploadProxy): {},
			utils.UDPPortOrDie(utils.PortDNS):         {},
		},
	}, &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: kp.RandomPorts,
		PortBindings:    portMap,
		ExtraHosts: []string{
			"nfs:192.168.66.2",
			"registry:192.168.66.2",
			"ceph:192.168.66.2",
		},
		Mounts: dnsmasqMounts,
	}, nil, nil, kp.Version+"-dnsmasq")

	if err := kp.Docker.ContainerStart(ctx, dnsmasq.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return dnsmasq.ID, nil
}

func (kp *KubevirtProvider) RunRegistry(ctx context.Context) (string, error) {
	err := docker.ImagePull(kp.Docker, ctx, utils.DockerRegistryImage, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	registry, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: utils.DockerRegistryImage,
	}, &container.HostConfig{
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.Version+"-registry")
	if err != nil {
		return "", err
	}

	if err := kp.Docker.ContainerStart(ctx, registry.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return registry.ID, nil
}

func (kp *KubevirtProvider) RunNFSGanesha(ctx context.Context) (string, error) {
	nfsData, err := filepath.Abs(kp.NFSData)
	if err != nil {
		return "", err
	}
	// Pull the ganesha image
	err = docker.ImagePull(kp.Docker, ctx, utils.NFSGaneshaImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	// Start the ganesha image
	nfsGanesha, err := kp.Docker.ContainerCreate(ctx, &container.Config{
		Image: utils.NFSGaneshaImage,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: nfsData,
				Target: "/data/nfs",
			},
		},
		Privileged:  true,
		NetworkMode: container.NetworkMode("container:" + kp.DNSMasq),
	}, nil, nil, kp.Version+"-nfs-ganesha")
	if err != nil {
		return "", err
	}

	if err := kp.Docker.ContainerStart(ctx, nfsGanesha.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	return nfsGanesha.ID, nil
}

func (kp *KubevirtProvider) Stop() {}

func (kp *KubevirtProvider) handleInterrupt(cancel context.CancelFunc, stop chan error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	cancel()
	stop <- fmt.Errorf("Interrupt received, clean up")
}
