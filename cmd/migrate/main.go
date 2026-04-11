package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/blagoySimandov/takgo/internal/adapters/sqlite"
	"github.com/blagoySimandov/takgo/internal/migrations"
	"github.com/uptrace/bun/migrate"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down>")
		os.Exit(1)
	}

	ctx := context.Background()

	database, err := sqlite.Open("takgo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	migrator := migrate.NewMigrator(database, migrations.Migrations)
	if err := migrator.Init(ctx); err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "up":
		if err := migrateUp(ctx, migrator); err != nil {
			log.Fatal(err)
		}
	case "down":
		if err := migrateDown(ctx, migrator); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func migrateUp(ctx context.Context, migrator *migrate.Migrator) error {
	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Println("no migrations to run")
		return nil
	}
	fmt.Printf("migrated up to group %s\n", group)
	return nil
}

func migrateDown(ctx context.Context, migrator *migrate.Migrator) error {
	group, err := migrator.Rollback(ctx)
	if err != nil {
		return err
	}
	if group.IsZero() {
		fmt.Println("nothing to roll back")
		return nil
	}
	fmt.Printf("rolled back group %s\n", group)
	return nil
}
