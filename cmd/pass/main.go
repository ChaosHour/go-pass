package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ChaosHour/go-pass/internal/config"
	"github.com/ChaosHour/go-pass/internal/database"
	"github.com/fatih/color"
)

var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

func main() {
	cfg := config.ParseFlags()

	if cfg.Help {
		printHelp()
		os.Exit(0)
	}

	// Check if ~/.my.cnf exists
	home := os.Getenv("HOME")
	if home == "" {
		log.Fatal(red("[!]"), "HOME environment variable not set")
	}
	if _, err := os.Stat(home + "/.my.cnf"); os.IsNotExist(err) {
		fmt.Println(red("[!]"), "Please create a ~/.my.cnf file with the database credentials.")
		os.Exit(1)
	}

	if err := cfg.LoadMyCnf(); err != nil {
		log.Fatal(red("[!]"), err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal(red("[!]"), err)
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		log.Fatal(red("[!]"), err)
	}
	defer db.Close()

	if err := database.DumpUserAccounts(ctx, db, cfg); err != nil {
		log.Fatal(red("[!]"), err)
	}

	// Sleep for 5 seconds as in original
	time.Sleep(5 * time.Second)

	if err := database.RunQuery(ctx, db, cfg); err != nil {
		log.Fatal(red("[!]"), err)
	}

	log.Println(green("[+]"), "Operation completed successfully")
}

func printHelp() {
	fmt.Println("Usage: go-pass -s <source host> -f <dump file>")
	fmt.Println("Options:")
	fmt.Println("  -s <source host>  Source MySQL host")
	fmt.Println("  -f <dump file>    Output dump file")
	fmt.Println("  -o <user>         Only dump the specified user")
	fmt.Println("  --format <fmt>    Output format: raw, import, pt-like (default: raw)")
	fmt.Println("  -h                Print this help")
}
