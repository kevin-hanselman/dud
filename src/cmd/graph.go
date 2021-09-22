package cmd

import (
	"github.com/awalterschulze/gographviz"
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
upstream stages in Graphviz DOT format. If no stage files are passed in, graph
will act on all stages in the index.

You can pipe the output of this command to 'dot' from the graphviz package to
generate images of the stage graph. Visit https://graphviz.org for more
information about Graphviz and for installation instructions.`,
	Example: "dud graph | dot -Tpng -o dud.png",
	Run: func(cmd *cobra.Command, paths []string) {
		_, _, idx, err := prepare(paths...)
		if err != nil {
			fatal(err)
		}

		if len(idx) == 0 {
			fatal(emptyIndexError{})
		}

		if len(paths) == 0 { // By default, run on the entire Index
			for path := range idx {
				paths = append(paths, path)
			}
		}

		graph := gographviz.NewEscape()
		if err := graph.SetDir(true); err != nil {
			fatal(err)
		}
		for _, path := range paths {
			inProgress := make(map[string]bool)
			if err := idx.Graph(path, inProgress, graph, onlyStages); err != nil {
				fatal(err)
			}
		}
		logger.Info.Println(graph.String())
	},
}
