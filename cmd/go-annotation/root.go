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
	"strings"

	"github.com/lucasew/go-annotation/annotation"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "go-annotation [folder|config.yaml]",
	Short: "Quickly make image annotations",
	Long: strings.TrimSpace(`
With a set of trivial choices scale the classification of a set of images to many people to build datasets to train classifiers.
    `),
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

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("Error executing command: %v", err)
		os.Exit(1)
	}
}

func init() {
	// Optional flags (only used when not providing a folder argument)
	rootCmd.Flags().StringP("config", "c", "", "Config file for the annotation")
	rootCmd.Flags().StringP("database", "d", "", "Database file path")
	rootCmd.Flags().StringP("images", "i", "", "Images directory path")
	rootCmd.Flags().StringP("addr", "a", ":8080", "Address to bind the webserver")
}
