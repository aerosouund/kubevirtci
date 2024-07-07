package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindbase"
	"kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/sriov"
)

var kindProvider kind.KindProvider

func NewRunKindCommand() *cobra.Command {
	rk := &cobra.Command{
		Use:   "run-kind",
		Short: "runs a kind provider",
		RunE:  runKind,
		Args:  cobra.ExactArgs(1),
	}
	rk.Flags().UintP("nodes", "n", 1, "number of cluster nodes to start")
	return rk
}

func runKind(cmd *cobra.Command, args []string) error {
	nodes, err := cmd.Flags().GetUint("nodes")
	if err != nil {
		return err
	}
	kindVersion := args[0]
	conf := &kind.KindConfig{
		Nodes:   int(nodes),
		Version: kindVersion,
	}

	switch kindVersion {
	case "k8s-1.28":
		kindProvider, err = kind.NewKindBaseProvider(conf)
		if err != nil {
			return err
		}
	case "sriov":
		kindProvider, err = sriov.NewKindSriovProvider(conf)
		if err != nil {
			return err
		}
	case "ovn":
	case "vgpu":
	default:
		return fmt.Errorf("Invalid k8s version passed, please use one of k8s-1.28, sriov, ovn or vgpu")
	}

	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	err = kindProvider.Start(ctx, cancel)
	if err != nil {
		return err
	}
	return nil
}
