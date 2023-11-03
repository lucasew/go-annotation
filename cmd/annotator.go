/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/lucasew/go-annotation/annotation"
	"github.com/spf13/cobra"
)

// annotatorCmd represents the annotator command
var annotatorCmd = &cobra.Command{
	Use:   "annotator",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
        configFile, err := cmd.Flags().GetString("config")
        if err != nil {
            panic(err)
        }
        config, err := annotation.LoadConfig(configFile)
        if err != nil {
            panic(err)
        }
        spew.Dump(config)
		fmt.Println("annotator called")
	},
}

func init() {
	rootCmd.AddCommand(annotatorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	annotatorCmd.PersistentFlags().StringP("config", "c", "", "Config file for the annotation")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// annotatorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
