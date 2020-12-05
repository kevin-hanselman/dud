package index

import (
	"fmt"

	"github.com/awalterschulze/gographviz"

	"github.com/pkg/errors"
)

// Graph creates a dependency graph starting from the given Stage.
func (idx Index) Graph(
	stagePath string,
	inProgress map[string]bool,
	graph *gographviz.Escape,
	onlyStages bool,
) error {
	// A subgraph MUST start with "cluster" for its "label" attribute to be displayed.
	// Intuitive, I know.
	// See: https://stackoverflow.com/a/7586857/857893
	stageSubgraphName := "cluster_" + stagePath
	if onlyStages {
		if graph.IsNode(stagePath) {
			return nil
		}
	} else {
		if graph.IsSubGraph(stageSubgraphName) {
			return nil
		}
	}

	// If we've visited this Stage but haven't recorded its status (the check
	// above), then we're in a cycle.
	if inProgress[stagePath] {
		return errors.New("cycle detected")
	}
	inProgress[stagePath] = true

	en, ok := idx[stagePath]
	if !ok {
		return fmt.Errorf("status: unknown stage %#v", stagePath)
	}

	graph.SetDir(true) // Ensure the graph is directed.
	graph.AddAttr(graph.Name, "rankdir", "LR")
	for artPath := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			return err
		}
		hasOwner := ownerPath != ""
		// If we're drawing the full graph, always draw an edge to the
		// dependency Artifact. Otherwise, draw an edge to the owner Stage if
		// one exists.
		if !onlyStages {
			if !graph.IsNode(artPath) {
				graph.AddNode(graph.Name, artPath, nil)
			}
			if err := graph.AddEdge(stageSubgraphName, artPath, true, nil); err != nil {
				return err
			}
		} else if hasOwner {
			if err := graph.AddEdge(stagePath, ownerPath, true, nil); err != nil {
				return err
			}
		}
		if hasOwner {
			if err := idx.Graph(ownerPath, inProgress, graph, onlyStages); err != nil {
				return err
			}
		}
	}
	if onlyStages {
		if err := graph.AddNode(graph.Name, stagePath, nil); err != nil {
			return err
		}
	} else {
		for artPath := range en.Stage.Outputs {
			if err := graph.AddNode(stageSubgraphName, artPath, nil); err != nil {
				return err
			}
		}
		err := graph.AddSubGraph(
			graph.Name,
			stageSubgraphName,
			map[string]string{"label": stagePath},
		)
		if err != nil {
			return err
		}
	}
	delete(inProgress, stagePath)
	return nil
}
