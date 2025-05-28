package cli

import (
	"fmt"

	"github.com/faelmori/cleandgo"
	"github.com/spf13/cobra"

	gl "github.com/faelmori/cleandgo/logger"
	vs "github.com/faelmori/cleandgo/version"
)

func ParserCmdList() []*cobra.Command {
	return []*cobra.Command{
		parseCommand(),
	}
}

func parseCommand() *cobra.Command {
	var treeFileSource, composerTargetPath string
	var printTree bool
	var debug, onlyDirectories, onlyFiles, quiet bool

	var parseCmd = &cobra.Command{
		Use: "parse",
		Annotations: GetDescriptions([]string{
			"Parse a tree view file and generate all files and directories structure",
			"This command is used to parse a tree view file and generate all files and directories structure from a visual representation",
		}, false),
		Version: vs.GetVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			ft, ftErr := cleandgo.NewFileTree(treeFileSource, composerTargetPath, printTree, nil, debug)
			if ftErr != nil {
				gl.Log("error", fmt.Sprintf("Failed to create file tree: %s", ftErr))
				return
			}
			if err := ft.ParseTree(); err != nil {
				gl.Log("error", fmt.Sprintf("Failed to parse tree: %s", err))
				return
			}
			gl.Log("success", "Tree parsed successfully!!!")
			gl.Log("info", "See you later...")
		},
	}

	parseCmd.Flags().StringVarP(&treeFileSource, "source", "s", "", "Path to the tree view file")
	parseCmd.Flags().StringVarP(&composerTargetPath, "composer", "c", "", "Path to the composer target directory")
	parseCmd.Flags().BoolVarP(&printTree, "print", "p", false, "Print the tree view")
	parseCmd.Flags().BoolVarP(&onlyDirectories, "onlyDirectories", "D", false, "Only include directories in the output")
	parseCmd.Flags().BoolVarP(&onlyFiles, "onlyFiles", "F", false, "Only include files in the output")
	parseCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	parseCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress output messages")

	return parseCmd
}
