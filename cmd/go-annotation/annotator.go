/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lucasew/go-annotation/annotation"
	"github.com/spf13/cobra"
)

// annotatorCmd represents the annotator command
var annotatorCmd = &cobra.Command{
	Use:   "annotator [folder|config.yaml]",
	Short: "Start the annotation web server",
	Long: `Start the annotation web server.

If you provide a folder path, it will:
  - Create a default config.yaml in that folder
  - Create an annotations.db in that folder
  - Use the folder as the images directory

If you provide a config file path, it will:
  - Use that config to start the server
  - Require --database and --images flags

Examples:
  # Initialize and start from a folder
  go-annotation annotator ./my-project

  # Start with explicit config
  go-annotation annotator -c config.yaml -d annotations.db -i ./images
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configFile, databaseFile, imagesDir string
		var err error

		// Check if a positional argument was provided
		if len(args) == 1 {
			arg := args[0]

			// Check if it's a directory
			if stat, err := os.Stat(arg); err == nil && stat.IsDir() {
				// It's a folder - initialize it
				log.Printf("Detected folder argument: %s", arg)
				log.Printf("Initializing project in folder...")

				configFile = filepath.Join(arg, "config.yaml")
				databaseFile = filepath.Join(arg, "annotations.db")
				imagesDir = arg

				// Create config if it doesn't exist
				if _, err := os.Stat(configFile); os.IsNotExist(err) {
					log.Printf("Creating default config: %s", configFile)
					if err := createSampleConfig(configFile, imagesDir); err != nil {
						return fmt.Errorf("failed to create config: %w", err)
					}
				}
			} else {
				// Assume it's a config file
				configFile = arg

				// Require other flags
				databaseFile, err = cmd.Flags().GetString("database")
				if err != nil || databaseFile == "" {
					return fmt.Errorf("when providing a config file, --database flag is required")
				}

				imagesDir, err = cmd.Flags().GetString("images")
				if err != nil || imagesDir == "" {
					return fmt.Errorf("when providing a config file, --images flag is required")
				}
			}
		} else {
			// No positional arg - use flags
			configFile, err = cmd.Flags().GetString("config")
			if err != nil || configFile == "" {
				return fmt.Errorf("either provide a folder/config argument or use --config flag")
			}

			databaseFile, err = cmd.Flags().GetString("database")
			if err != nil || databaseFile == "" {
				return fmt.Errorf("--database flag is required")
			}

			imagesDir, err = cmd.Flags().GetString("images")
			if err != nil || imagesDir == "" {
				return fmt.Errorf("--images flag is required")
			}
		}

		// Load config
		config, err := annotation.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Open database
		db, err := annotation.GetDatabase(databaseFile)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()

		// Create app
		app := &annotation.AnnotatorApp{
			ImagesDir: imagesDir,
			Database:  db,
			Config:    config,
		}

		// Prepare database
		if err := app.PrepareDatabase(cmd.Context()); err != nil {
			return fmt.Errorf("failed to prepare database: %w", err)
		}

		// Get bind address
		addr, _ := cmd.Flags().GetString("addr")

		log.Printf("Configuration: %s", configFile)
		log.Printf("Database: %s", databaseFile)
		log.Printf("Images: %s", imagesDir)
		log.Printf("Tasks configured: %d", len(config.Tasks))
		for _, task := range config.Tasks {
			log.Printf("  - %s: %s", task.ID, task.Name)
		}
		log.Printf("Starting server on: %s", addr)

		return http.ListenAndServe(addr, app.GetHTTPHandler())
	},
}

func init() {
	rootCmd.AddCommand(annotatorCmd)

	// Optional flags (only used when not providing a folder argument)
	annotatorCmd.Flags().StringP("config", "c", "", "Config file for the annotation")
	annotatorCmd.Flags().StringP("database", "d", "", "Database file path")
	annotatorCmd.Flags().StringP("images", "i", "", "Images directory path")
	annotatorCmd.Flags().StringP("addr", "a", ":8080", "Address to bind the webserver")
}
