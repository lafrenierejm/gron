/*
Copyright © 2016 Tom Hudson <mail@tomnomnom.com>
Copyright © 2023 Joseph LaFreniere <git@lafreniere.xyz>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	internal "github.com/lafrenierejm/gron/internal/gron"
	"github.com/mattn/go-colorable"
	"github.com/spf13/cobra"
	json "github.com/virtuald/go-ordered-json"
)

var version = "0.0.1"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "gron",
	Version: version,
	Short:   "Transform JSON or YAML into discrete assignments to make it greppable",
	Long: `gron transforms JSON or YAML (from a file, URL, or stdin) into discrete assignments to make it easier to grep for what you want and see the absolute "path" to it.

Examples:
  gron /tmp/apiresponse.json
  gron http://jsonplaceholder.typicode.com/users/1
  curl -s http://jsonplaceholder.typicode.com/users/1 | gron
  gron http://jsonplaceholder.typicode.com/users/1 | grep company | gron --ungron
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		colorizeFlag, err := cmd.Flags().GetBool("colorize")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		insecureFlag, err := cmd.Flags().GetBool("insecure")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		jsonFlag, err := cmd.Flags().GetBool("json")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		monochromeFlag, err := cmd.Flags().GetBool("monochrome")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		sortFlag, err := cmd.Flags().GetBool("sort")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		streamFlag, err := cmd.Flags().GetBool("stream")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		ungronFlag, err := cmd.Flags().GetBool("ungron")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		valuesFlag, err := cmd.Flags().GetBool("values")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
		yamlFlag, err := cmd.Flags().GetBool("yaml")
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		var rawInput io.Reader
		if len(args) == 0 || args[0] == "" || args[0] == "-" {
			rawInput = os.Stdin
		} else {
			filename := args[0]
			if validURL(filename) {
				rawInput, err = getURL(filename, insecureFlag)
				if err != nil {
					log.Println(err)
					os.Exit(1)
				}
			} else {
				rawInput, err = os.Open(filename)
				if err != nil {
					log.Println(err)
					os.Exit(1)
				}
			}
		}

		var conv internal.StatementConv = internal.StatementToString
		var colorize bool = false
		if colorizeFlag {
			colorize = true
		} else if !monochromeFlag {
			nocolorEnv, nocolorEnvPresent := os.LookupEnv("NO_COLOR")
			if nocolorEnvPresent && nocolorEnv != "" {
				colorize = false
			} else {
				colorize = true
			}
		}
		if colorize {
			conv = internal.StatementToColorString
		}

		var actionExit int
		var actionErr error
		if ungronFlag {
			actionExit, actionErr = internal.Ungron(
				rawInput,
				colorable.NewColorableStdout(),
				jsonFlag,
				colorize,
			)
		} else if valuesFlag {
			actionExit, actionErr = gronValues(rawInput, colorable.NewColorableStdout())
		} else if streamFlag {
			actionExit, actionErr = internal.GronStream(
				rawInput,
				colorable.NewColorableStdout(),
				conv,
				yamlFlag,
				sortFlag,
				jsonFlag,
			)
		} else {
			actionExit, actionErr = internal.Gron(
				rawInput,
				colorable.NewColorableStdout(),
				conv,
				yamlFlag,
				sortFlag,
				jsonFlag,
			)
		}

		if actionExit != 0 || actionErr != nil {
			log.Println(err)
		}
		os.Exit(actionExit)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("colorize", "c", false, "Colorize output (default on TTY)")
	rootCmd.Flags().BoolP("insecure", "k", false, "Disable certificate validation when reading from a URL")
	rootCmd.Flags().BoolP("json", "j", false, "Represent gron data as JSON stream")
	rootCmd.Flags().BoolP("monochrome", "m", false, "Do not colorize output")
	rootCmd.Flags().BoolP("sort", "", false, "Sort output")
	rootCmd.Flags().BoolP("stream", "s", false, "Treat each line of input as a separate JSON object")
	rootCmd.Flags().BoolP("ungron", "u", false, "Reverse the operation (turn assignments back into JSON)")
	rootCmd.Flags().BoolP("values", "v", false, "Print just the values of provided assignments")
	rootCmd.Flags().BoolP("version", "", false, "Print version information")
	rootCmd.Flags().BoolP("yaml", "y", false, "Treat input as YAML instead of JSON")
}

// gronValues prints just the scalar values from some input gron statements
// without any quotes or anything of that sort; a bit like jq -r
// e.g. json[0].user.name = "Sam"; -> Sam
func gronValues(r io.Reader, w io.Writer) (int, error) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		s := internal.StatementFromString(scanner.Text())

		// strip off the leading 'json' bare key
		if s[0].Typ == internal.TypBare && s[0].Text == "json" {
			s = s[1:]
		}

		// strip off the leading dots
		if s[0].Typ == internal.TypDot || s[0].Typ == internal.TypLBrace {
			s = s[1:]
		}

		for _, t := range s {
			switch t.Typ {
			case internal.TypString:
				var text string
				err := json.Unmarshal([]byte(t.Text), &text)
				if err != nil {
					// just swallow errors and try to continue
					continue
				}
				fmt.Println(text)

			case internal.TypNumber, internal.TypTrue, internal.TypFalse, internal.TypNull:
				fmt.Println(t.Text)

			default:
				// Nothing
			}
		}
	}

	return 0, nil
}
