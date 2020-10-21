package index

import (
	"fmt"

	"github.com/awalterschulze/gographviz"

	"github.com/pkg/errors"
)

// Use an undirected edge to differentiate ownership from a dependency. Faking
// undirected edges in a directed graph requires the attribute "dir=none".
// See: https://stackoverflow.com/questions/13236975
var ownershipEdge = map[string]string{"dir": "none"}
var stageNode = map[string]string{"color": "cadetblue1", "style": "filled"}
var artifactNode = map[string]string{"shape": "box"}

// Graph creates a dependency graph starting from the given Stage.
func (idx Index) Graph(
	stagePath string,
	inProgress map[string]bool,
	graph *gographviz.Escape,
	onlyStages bool,
) error {
	if graph.IsNode(stagePath) {
		return nil
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
	for artPath := range en.Stage.Dependencies {
		ownerPath, _, err := idx.findOwner(artPath)
		if err != nil {
			errors.Wrap(err, "status")
		}
		hasOwner := ownerPath != ""
		// If we're drawing the full graph, always draw an edge to the
		// dependency Artifact. Otherwise, draw an edge to the owner Stage if
		// one exists.
		if !onlyStages {
			if !graph.IsNode(artPath) {
				graph.AddNode(graph.Name, artPath, artifactNode)
			}
			if err := graph.AddEdge(stagePath, artPath, true, nil); err != nil {
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
	if !onlyStages {
		for artPath := range en.Stage.Outputs {
			graph.AddNode(graph.Name, artPath, artifactNode)
			if err := graph.AddEdge(stagePath, artPath, true, ownershipEdge); err != nil {
				return err
			}
		}
	}
	graph.AddNode(graph.Name, stagePath, stageNode)
	delete(inProgress, stagePath)
	return nil
}
