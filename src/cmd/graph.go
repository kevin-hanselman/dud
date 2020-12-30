package cmd

import (
	"fmt"
	"os"

	"github.com/awalterschulze/gographviz"
	"github.com/kevin-hanselman/dud/src/index"
	"github.com/spf13/cobra"
)

var onlyStages bool

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().BoolVar(
		&onlyStages,
		"stages-only",
		false,
		"only show stages; no artifacts",
	)
}

var graphCmd = &cobra.Command{
	Use:   "graph [flags] [stage_file]...",
	Short: "Print the stage graph in graphviz DOT format",
	Long: `Graph prints the stage graph in graphviz DOT format.

For each stage file passed in, graph will print the graph of the stage and all
upstream stages. If no stage files are passed in, graph will act on all stages
in the index.

You can pipe the output of this command to 'dot' from the graphviz package to
generate images of the stage graph.`,
	Example: "dud graph | dot -Tpng -o dud.png",
	PreRun:  requireInitializedProject,
	Run: func(cmd *cobra.Command, args []string) {
		idx, err := index.FromFile(".dud/index")
		if os.IsNotExist(err) { // TODO: print error instead?
			idx = make(index.Index)
		} else if err != nil {
			logger.Fatal(err)
		}

		if len(args) == 0 { // By default, run on the entire Index
			for path := range idx {
				args = append(args, path)
			}
		}

		graph := gographviz.NewEscape()
		graph.SetDir(true)
		for _, path := range args {
			inProgress := make(map[string]bool)
			if err := idx.Graph(path, inProgress, graph, onlyStages); err != nil {
				logger.Fatal(err)
			}
		}
		fmt.Println(graph.String())
	},
}
