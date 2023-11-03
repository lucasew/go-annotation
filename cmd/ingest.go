/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
    "log"
    "io/fs"

	"image"
	"path/filepath"

	"github.com/lucasew/go-annotation/annotation"
	"github.com/spf13/cobra"
)

// ingestCmd represents the ingest command
var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest a folder of files to a folder of images.",
	Long: `Ingest a folder of files that were extracted from somewhere and organize in a flat hierarchy of images.`,
    Args: func (cmd *cobra.Command, args []string) error {
        if err := cobra.MinimumNArgs(2)(cmd, args); err != nil {
            return err
        }
        inputs := args[0:len(args) - 1]
        output := args[len(args)-1]
        for i, input := range(inputs) {
            fileInfo, err := os.Stat(input)
            if err != nil {
                return fmt.Errorf("on %dth argument: %w", i + 1, err)
            }
            if !fileInfo.IsDir() {
                return fmt.Errorf("on %dth argument: must be a directory", i + 1)
            }
        }
        return os.MkdirAll(output, 0777)
    },
	Run: func(cmd *cobra.Command, args []string) {
        inputs := args[0:len(args) - 1]
        output := args[len(args)-1]

        crawledFilepaths := make(chan image.Image, 10) // pipeline

        ingestWorker := func(queue chan image.Image) {
            for image := range queue {
                err := annotation.IngestImage(image, output)
                if err != nil {
                    log.Printf("Ingesting image error: %s", err)
                }
            }
        }

        go ingestWorker(crawledFilepaths)

        for _, input := range inputs {
            filepath.WalkDir(input, func(path string, info fs.DirEntry, err error) error {
                if err != nil {
                    return err
                }
                if info.IsDir() {
                    return nil
                }
                img, err := annotation.DecodeImage(path)
                if err != nil {
                    return nil
                }
                log.Printf("found image '%s'", path)
                crawledFilepaths <- img
                return nil
            })
        }
	},
}

func init() {
	rootCmd.AddCommand(ingestCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ingestCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ingestCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
