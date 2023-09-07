package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const content = `keep_dirs: true
#deps:
#- url: https://github.com/audiotool/manual
#  from: images/
#  to: my/local/path/
#  include: pulverisateur.\.png # regex
#  exclude: # regex
#  options:
#    ref: "branchname" # can be "heads/<branchname>" "tags/<tagname>" "commit/<commitsha>" or simply "<name>"
`

// createCmd represents the batch command
var createCmd = &cobra.Command{
	Use:   "init",
	Short: "init creates a new pasta file in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		// check if file already exists
		_, err := os.Stat(pastayaml)

		if err == nil {
			fmt.Fprintf(os.Stderr, "'%s' already exists\n", pastayaml)
			os.Exit(-1)
		}

		// drop a warning if the parent directory has a file
		pathToYaml, err := findPastaFile()

		if err == nil {
			fmt.Printf("Warning: Found a pasta file on '%s'\n", pathToYaml)
		}

		// write file
		err = os.WriteFile(pastayaml, []byte(content), 0664)

		if err != nil {
			fmt.Fprintf(os.Stderr, "could not create pasta file: %v\n", err)
			os.Exit(-2)
		}

		fmt.Printf("Created '%s'\n", pastayaml)
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
}
