package cmd

import (
	"github.com/spf13/cobra"
	kind "kubevirt.io/kubevirtci/cluster-provision/gocli/providers/kind/kindcommon"
)

func NewRemoveKindCommand() *cobra.Command {

	rm := &cobra.Command{
		Use:   "rm",
		Short: "rm deletes all traces of a cluster",
		RunE:  rmKind,
		Args:  cobra.ExactArgs(1),
	}
	return rm
}

func rmKind(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	kindProvider, err := kind.NewKindCommondProvider(prefix, 0)
	if err != nil {
		return err
	}
	if err = kindProvider.Delete(); err != nil {
		return err
	}

	return nil
}
