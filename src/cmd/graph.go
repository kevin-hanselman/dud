package cmd

import (
	"fmt"
	"log"
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
		"don't show artifacts in the graph",
	)
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Print Stage graph in graphviz DOT format",
	Long:  "Print Stage graph in graphviz DOT format",
	Run: func(cmd *cobra.Command, args []string) {
		idx, err := index.FromFile(".dud/index")
		if os.IsNotExist(err) { // TODO: print error instead?
			idx = make(index.Index)
		} else if err != nil {
			log.Fatal(err)
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
				log.Fatal(err)
			}
		}
		fmt.Println(graph.String())
	},
}
