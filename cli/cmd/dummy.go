package cmd

import (
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var dummyCmd = &cobra.Command{
	Use:   "dummy",
	Short: "dummy task for tests",
	Run: func(cmd *cobra.Command, args []string) {

		// model := pkg.BuildModelFromFolder("./")

		// model.MustNoErrors()

		// model.UpsertVar("main.hcl", "foo", "\"hello world xx\"")
		// model.MustNoErrors()

		// // fmt.Println("run called")

	},
}

func init() {
	rootCmd.AddCommand(dummyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
