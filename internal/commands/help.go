package commands

import "fmt"

func HandleHelp(args []string) error {
	help := `dredge - Encrypted storage for secrets, credentials, and config files

Usage:
  dredge <command> [arguments]

Commands:
  add, a      Add a new item
  search, s   Search for items
  view, v     View an item by ID
  edit, e     Edit an item
  rm          Remove an item
  link, ln    Link an item to a system path
  unlink      Unlink an item from system path
  init        Initialize git repository for sync
  push        Push changes to remote
  pull        Pull changes from remote
  sync        Sync with remote (pull + push)
  help, h     Show this help message

Examples:
  dredge add ssh-config My SSH Config -c "Host github.com..." #ssh #config
  dredge search ssh
  dredge view ssh-config
  dredge link ssh-config ~/.ssh/config
`
	fmt.Print(help)
	return nil
}
