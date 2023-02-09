package toolbox

import (
	"github.com/spf13/cobra"
)

func (self *toolboxImpl) newRootParser() *cobra.Command {
	root := &cobra.Command{
		Use:   "sre",
		Short: "SRE cloud toolbox and command line",
		Long: "The SRE command line which is used to interact with " +
			"resource of the whole company",
	}
	root.AddCommand(self.newBizflyParser())
	root.AddCommand(self.newSettingParser())
	root.AddCommand(self.newHealthParser())
	root.AddCommand(self.newKubernetesParser())
	return root
}
