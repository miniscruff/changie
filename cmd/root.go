package cmd

import "github.com/spf13/cobra"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "changie",
	Short: "changie handles conflict-free changelog management",
	Long: `Changie keeps your changes organized and attached to your code.

Changie is aimed at seemlessly integrating into your release process while also
being easy to use for developers and your release team.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) error {
	rootCmd.Version = version
	return rootCmd.Execute()
}
