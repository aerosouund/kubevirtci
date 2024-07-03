package cmd

import (
	"context"

	"github.com/spf13/cobra"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kindcommon"
)

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
	k8sVersion := args[0]

	kindProvider, err := kind.NewKindCommondProvider(k8sVersion, int(nodes))
	if err != nil {
		return err
	}
	b := context.Background()
	ctx, cancel := context.WithCancel(b)

	err = kindProvider.Start(ctx, cancel)
	if err != nil {
		return err
	}
	return nil
}
