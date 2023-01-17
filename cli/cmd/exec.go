/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"strings"

	"github.com/v-braun/ratt/pkg"

	"github.com/spf13/cobra"
)

var execArgsP *[]string

// runCmd represents the run command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Executes all passed request/testcase",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rawExecArgs := *execArgsP

		execArgs := map[string]string{}
		for _, arg := range rawExecArgs {
			assignIdx := strings.Index(arg, "=")
			if assignIdx <= 0 {
				continue
			}
			k := arg[0:assignIdx]
			v := arg[assignIdx+1:]

			execArgs[k] = v
		}

		model := pkg.BuildModelFromFolder("./")
		model.MustNoErrors()
		model.Exec(args[0], execArgs)
		model.MustNoErrors()

		// fmt.Println("run called")
	},
}

func init() {

	execArgsP = execCmd.Flags().StringArrayP("arg", "a", []string{}, "args to be passed to the request/testcase")
	rootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
