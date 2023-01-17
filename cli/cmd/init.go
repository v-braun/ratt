/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const repoFileName string = "repository.hcl"
const repoFileContent string = `
declare "request" "create_user" {  
	args{
	  firstName = ""
	  lastName = ""
	}
  
	url = "${vars.baseUrl}/user"
	method = "POST"
	body = <<EOT
	  {
		"firstName": "${args.firstName}",
		"lastName": "${args.lastName}"
	}
	EOT
}

declare "request" "get_user" {  
	args{
		id = ""
	}
  
	url = "${vars.baseUrl}/user/${args.id}"
	method = "GET"
}
`

const mainFileName string = "main.hcl"
const mainFileContent string = `
vars {
	// base url to your API
	baseUrl = "my api here ..."	

	// variable to to transfer infirmation between requests
	newUserId = 0
}

invoke "request" "create_user"{
	baseUrl = vars.baseUrl
	firstName = "Ralph"
	lastName = "Boner"
  
	then "assert" {
	  status_to_be_201 = (res.statusCode == 201)
	  server_is_not_nginx = (res.headers["Server"] != "nginx")
	}
	then "set" {
	  // set the id of the new user to variable newUserId
	  newUserId = res.body.id
	}
	then "assert" {
	  status_to_be_200 = (res.statusCode == 201)
	}
}


invoke "request" "get_user"{
	id = "${vars.newUserId}"
  
	then "assert" {
	  status_to_be_200 = (res.statusCode == 200)
	  user_name_is_paul = (res.body.firstName == "Ralph")
	}
}    
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new RATT project",
	Long: `Scaffolds a new projects within the current directory
This command creates the following files:
- repository.hcl (contains vars and declarations)
- main.hcl  (all invkoe statements and variable declarations)`,
	Run: func(cmd *cobra.Command, args []string) {
		writeInitFile(repoFileName, repoFileContent)
		writeInitFile(mainFileName, mainFileContent)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func writeInitFile(fileName string, content string) {
	path := fmt.Sprintf("./%s", fileName)
	if fileExists(path) {
		fmt.Printf("%s %s\n", err("✖"), warn("file %s already exist, skip creation", fileName))
	} else {
		os.WriteFile(path, []byte(content), 0644)
		fmt.Printf("%s generated file %s\n", success("✔"), fileName)
	}
}
