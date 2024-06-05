package providers

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"
)

func NewKubevirtProvider(k8sversion string, options ...KubevirtProviderOption) *KubevirtProvider {
	bp := &KubevirtProvider{
		Version:    k8sversion,
		Nodes:      1,
		Numa:       1,
		Memory:     "3096M",
		CPU:        2,
		Background: true,
	}

	for _, option := range options {
		option(bp)
	}

	return bp
}

func (kp *KubevirtProvider) SetClient() {}

func (kp *KubevirtProvider) Start() (retErr error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	stop := make(chan error, 10)
	containers, _, done := docker.NewCleanupHandler(cli, stop, os.Stdout, false)

	defer func() {
		stop <- retErr
		<-done
	}()

	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		<-interrupt
		cancel()
		stop <- fmt.Errorf("Interrupt received, clean up")
	}()

	if containerSuffix != "" {
		kp.Version = fmt.Sprintf("%s/%s%s", kp.ContainerOrg, kp.Version, kp.ContainerSuffix)
	} else {
		kp.Version = path.Join(kp.ContainerOrg, kp.Version)
	}

	if kp.Slim {
		kp.Version += "-slim"
	}

	if len(kp.ContainerRegistry) > 0 {
		kp.Version = path.Join(kp.ContainerRegistry, kp.Version+":2403130317-a3e0778")
		fmt.Printf("Download the image %s\n", kp.Version)
		err = docker.ImagePull(cli, ctx, kp.Version, types.ImagePullOptions{})
		if err != nil {
			panic(fmt.Sprintf("Failed to download cluster image %s, %s", kp.Version, err))
		}
	}
	return nil
}

func (kp *KubevirtProvider) Stop() {}
