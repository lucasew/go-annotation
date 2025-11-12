/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	// "github.com/davecgh/go-spew/spew"
	"github.com/lucasew/go-annotation/annotation"
	"github.com/spf13/cobra"
)

func PrintQuery(ctx context.Context, db *sql.Tx, query string, args ...interface{}) error {
	stmt, err := db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	result, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	columns, err := result.Columns()
	if err != nil {
		return err
	}
	if len(columns) > 1 {
		fmt.Println(strings.Join(columns, "\t"))
	}
	pointers := make([]interface{}, len(columns))
	container := make([]string, len(columns))
	for i := 0; i < len(columns); i++ {
		pointers[i] = &container[i]
	}
	for result.Next() {
		result.Scan(pointers...)
		fmt.Println(strings.Join(container, "\t"))
	}
	return nil
}

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [flags] database step [value] [item]",
	Short: "Queries the annotation database",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fetchHash, err := cmd.Flags().GetBool("show-hashes")
		if err != nil {
			return err
		}
		if len(args) < 1 {
			return cmd.Help()
		}
		db, err := annotation.GetDatabase(args[0])
		if err != nil {
			return err
		}
		defer db.Close()
		// spew.Dump(db)

		tx, err := db.BeginTx(cmd.Context(), &sql.TxOptions{
			Isolation: sql.LevelReadUncommitted,
		})
		if err != nil {
			return err
		}
		defer tx.Rollback()

		queryArgs := []interface{}{}
		query := ""
		if len(args) < 2 {
			return PrintQuery(cmd.Context(), tx, "select distinct substr(tbl_name, 6) from SQLITE_MASTER where tbl_name like 'task_%'")
		}
		if len(args) < 3 {
			return PrintQuery(cmd.Context(), tx, fmt.Sprintf("select distinct value from task_%s where sure = 1", args[1]))
		}

		if fetchHash {
			query += "select image "
		} else {
			query += "select images.filename "
		}
		query += fmt.Sprintf(" from task_%s ", args[1])
		query += "join images on image = sha256 "
		query += "where sure = 1 "
		if len(args) >= 3 {
			query += " and value = ?"
			queryArgs = append(queryArgs, args[2])
		}
		if len(args) >= 4 {
			query += " and (image = ? or images.filename = ?)"
			queryArgs = append(queryArgs, args[3], args[3])
		}

		return PrintQuery(cmd.Context(), tx, query, queryArgs...)
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().BoolP("show-hashes", "i", false, "Show hash of file instead of file name")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// annotatorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
