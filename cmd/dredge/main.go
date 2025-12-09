package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/DeprecatedLuar/dredge/internal/commands"
	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

var (
	debugMode  bool
	luckMode   bool
	searchMode bool
)

func main() {
	app := &cli.App{
		Name:  "dredge",
		Usage: "Encrypted storage for secrets, credentials, and config files",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "Password for decryption (skips prompt)",
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "Enable debug output",
				Destination: &debugMode,
			},
			&cli.BoolFlag{
				Name:        "luck",
				Aliases:     []string{"l"},
				Usage:       "Force view top search result",
				Destination: &luckMode,
			},
			&cli.BoolFlag{
				Name:        "search",
				Aliases:     []string{"s"},
				Usage:       "Force show search list",
				Destination: &searchMode,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "add",
				Aliases: []string{"a", "new", "+"},
				Usage:   "Add a new item",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Usage: "Custom ID for the item",
					},
					&cli.StringFlag{
						Name:    "content",
						Aliases: []string{"c"},
						Usage:   "Content of the item",
					},
					&cli.StringSliceFlag{
						Name:    "tags",
						Aliases: []string{"t"},
						Usage:   "Tags for the item",
					},
				},
				Action: func(c *cli.Context) error {
					if c.NArg() < 1 && c.String("id") == "" {
						return fmt.Errorf("usage: dredge add [--id <id>] <title...> [-c <content...>] [-t <tag1> <tag2> ...]")
					}

					id := c.String("id")
					content := c.String("content")
					tags := c.StringSlice("tags")

					// Title is remaining args
					title := strings.Join(c.Args().Slice(), " ")

					if title == "" && id == "" {
						return fmt.Errorf("title is required")
					}

					return commands.HandleAdd(id, title, content, tags)
				},
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Usage:   "Search for items",
				Action: func(c *cli.Context) error {
					query := strings.Join(c.Args().Slice(), " ")
					return commands.HandleSearch(query, luckMode, searchMode)
				},
			},
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   "List all items",
				Action: func(c *cli.Context) error {
					return commands.HandleList(c.Args().Slice())
				},
			},
			{
				Name:    "view",
				Aliases: []string{"v"},
				Usage:   "View an item by ID",
				Action: func(c *cli.Context) error {
					return commands.HandleView(c.Args().Slice())
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit an item",
				Action: func(c *cli.Context) error {
					return commands.HandleEdit(c.Args().Slice())
				},
			},
			{
				Name:  "rm",
				Usage: "Remove an item",
				Action: func(c *cli.Context) error {
					return commands.HandleRemove(c.Args().Slice())
				},
			},
			{
				Name:  "undo",
				Usage: "Restore last deleted item",
				Action: func(c *cli.Context) error {
					return commands.HandleUndo(c.Args().Slice())
				},
			},
			{
				Name:    "link",
				Aliases: []string{"ln"},
				Usage:   "Link an item to a system path",
				Action: func(c *cli.Context) error {
					return commands.HandleLink(c.Args().Slice())
				},
			},
			{
				Name:  "unlink",
				Usage: "Unlink an item from system path",
				Action: func(c *cli.Context) error {
					return commands.HandleUnlink(c.Args().Slice())
				},
			},
			{
				Name:  "init",
				Usage: "Initialize git repository for sync",
				Action: func(c *cli.Context) error {
					return commands.HandleInit(c.Args().Slice())
				},
			},
			{
				Name:  "push",
				Usage: "Push changes to remote",
				Action: func(c *cli.Context) error {
					return commands.HandlePush(c.Args().Slice())
				},
			},
			{
				Name:  "pull",
				Usage: "Pull changes from remote",
				Action: func(c *cli.Context) error {
					return commands.HandlePull(c.Args().Slice())
				},
			},
			{
				Name:  "sync",
				Usage: "Sync with remote (pull + push)",
				Action: func(c *cli.Context) error {
					return commands.HandleSync(c.Args().Slice())
				},
			},
		},
		Before: func(c *cli.Context) error {
			// Cache password if provided via flag
			if password := c.String("password"); password != "" {
				Debugf("Caching password from --password flag")
				if err := crypto.CachePassword(password); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to cache password: %v\n", err)
				} else {
					Debugf("Password cached successfully")
				}
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			// Default action: smart query routing
			// Handles: dredge 1, dredge <id>, dredge <search-query>
			if c.NArg() == 0 {
				cli.ShowAppHelp(c)
				return nil
			}

			args := c.Args().Slice()
			firstArg := args[0]

			// Try as numbered result first (if single numeric arg)
			if len(args) == 1 {
				if num, err := strconv.Atoi(firstArg); err == nil && num > 0 {
					if id, cacheErr := storage.GetCachedResult(num); cacheErr == nil {
						return commands.HandleView([]string{id})
					}
					// If cache miss, fall through to try as ID/search
				}

				// Try as direct ID
				if err := commands.HandleView([]string{firstArg}); err == nil {
					return nil
				}
			}

			// Fall back to search
			query := strings.Join(args, " ")
			return commands.HandleSearch(query, luckMode, searchMode)
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func Debugf(format string, args ...any) {
	if debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}
