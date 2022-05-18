// Code generated. DO NOT EDIT.

package generated

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(completionCmd)
}

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Emits bash a completion for apg",
	Long: `Enable bash completion like so:
		Linux:
			source <(apg completion)
		Mac:
			brew install bash-completion
			apg completion > $(brew --prefix)/etc/bash_completion.d/apg`,
	Run: func(cmd *cobra.Command, args []string) {
		rootCmd.GenBashCompletion(os.Stdout)
	},
}
