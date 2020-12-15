package index

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/awalterschulze/gographviz"

	"github.com/pkg/errors"
)

var hiddenAttr = map[string]string{"style": "invis", "shape": "point"}

type stageNode struct {
	Path    string
	Command string
}

var stageTemplate string = `<
<table border="0">
<tr><td>{{ .Path }}</td></tr>
{{ if .Command }}
<hr/>
<tr><td>{{ .Command }}</td></tr>
{{ end }}</table>
>`

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

	// Ensure the graph is directed, and disallow multiple edges between the same nodes.
	graph.SetDir(true)
	graph.SetStrict(true)
	// Draw the graph left to right. When drawn top-down, graph edges tend to
	// be drawn through stage names.
	graph.AddAttr(graph.Name, "rankdir", "LR")
	// Must be true for edges to be directly connected to a subgraph.
	// See: https://stackoverflow.com/a/2012106/857893
	graph.AddAttr(graph.Name, "compound", "true")
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
			// Draw the edge from the subgraph (stage) to the Artifact
			// dependency. Unfortunately this requires serious chicanery.
			// First, compound=true needs to be set on the graph (see above).
			// Second, the subgraph must be set as the source side of the edge
			// with ltail (see below). Third, the edge's source node must be an
			// actual node in the subgraph, so we use a dummy node named after
			// the Stage.
			// See: https://stackoverflow.com/a/2012106/857893
			attrs := map[string]string{"ltail": stageSubgraphName}
			if err := graph.AddEdge(stagePath, artPath, true, attrs); err != nil {
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
		buf := bytes.Buffer{}
		tmpl, err := template.New("stage").Parse(stageTemplate)
		if err != nil {
			return err
		}
		tmpl.Execute(&buf, stageNode{Path: stagePath, Command: en.Stage.Command})
		if err := graph.AddSubGraph(
			graph.Name,
			stageSubgraphName,
			map[string]string{"label": buf.String()},
		); err != nil {
			return err
		}
		// Add a dummy node for drawing edges from a Stage to its dependencies. See above.
		if err := graph.AddNode(stageSubgraphName, stagePath, hiddenAttr); err != nil {
			return err
		}
	}
	delete(inProgress, stagePath)
	return nil
}
