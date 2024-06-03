package main

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/cmd/utils"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/docker"

	k8s "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/k8s"
	sshutils "kubevirt.io/kubevirtci/cluster-provision/gocli/utils/ssh"
)

func TestClusterRunning(t *testing.T) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	assert.NoError(t, err)

	containers, err := docker.GetPrefixedContainers(cli, "kubevirt-dnsmasq")
	assert.NoError(t, err)

	container, err := cli.ContainerInspect(context.Background(), containers[0].ID)
	assert.NoError(t, err)

	apiServerPort, err := utils.GetPublicPort(utils.PortAPI, container.NetworkSettings.Ports)
	assert.NoError(t, err)

	sshPort, err := utils.GetPublicPort(utils.PortSSH, container.NetworkSettings.Ports)
	assert.NoError(t, err)

	err = sshutils.CopyRemoteFile(sshPort, "/etc/kubernetes/admin.conf", ".kubeconfig")
	assert.NoError(t, err)

	config, err := k8s.InitConfig(".kubeconfig", apiServerPort)
	assert.NoError(t, err)

	k8sClient, err := k8s.NewDynamicClient(config)
	assert.NoError(t, err)

	_, err = k8sClient.List(schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespaces"}, "")
	assert.NoError(t, err)

}
