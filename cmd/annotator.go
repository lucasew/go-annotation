/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/lucasew/go-annotation/annotation"
    "log"
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
	RunE: func(cmd *cobra.Command, args []string) error {
        configFile, err := cmd.Flags().GetString("config")
        if err != nil { return err }

        config, err := annotation.LoadConfig(configFile)
        if err != nil { return err }

        spew.Dump(config)
        databaseFile, err := cmd.Flags().GetString("database")
        if err != nil { return err }
        db, err := annotation.GetDatabase(databaseFile)
        if err != nil { return err }
        defer db.Close()
        spew.Dump(db)

        imagesDir, err := cmd.Flags().GetString("images")
        if err != nil { return err }


        err = annotation.PrepareDatabase(cmd.Context(), db, config, imagesDir)
        if err != nil { return err }
        spew.Dump(databaseFile, imagesDir)
        for _, task := range config.Tasks {
            log.Printf("task: %s -- %s", task.ID, task.Name)
        }
        return nil
	},
}

func init() {
	rootCmd.AddCommand(annotatorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	annotatorCmd.PersistentFlags().StringP("config", "c", "", "Config file for the annotation")
    annotatorCmd.MarkPersistentFlagRequired("config")
    annotatorCmd.MarkPersistentFlagFilename("config")
	annotatorCmd.PersistentFlags().StringP("database", "d", "", "Where to store the annotation database")
    annotatorCmd.MarkPersistentFlagRequired("database")
    annotatorCmd.MarkPersistentFlagFilename("database")
	annotatorCmd.PersistentFlags().StringP("images", "i", "", "Where to store the images")
    annotatorCmd.MarkPersistentFlagDirname("images")
    annotatorCmd.MarkPersistentFlagRequired("images")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// annotatorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
