package index

import (
	"github.com/awalterschulze/gographviz"
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/duc/src/artifact"
	"github.com/kevin-hanselman/duc/src/stage"

	"testing"
)

func assertGraphsEqual(
	graphWant *gographviz.Escape,
	graphGot *gographviz.Escape,
	t *testing.T,
) {
	if diff := cmp.Diff(graphWant.Nodes.Sorted(), graphGot.Nodes.Sorted()); diff != "" {
		t.Fatalf("graph.Nodes -want +got:\n%s", diff)
	}
	if diff := cmp.Diff(graphWant.Edges.Sorted(), graphGot.Edges.Sorted()); diff != "" {
		t.Fatalf("graph.Edges -want +got:\n%s", diff)
	}
}

func TestGraph(t *testing.T) {

	t.Run("disjoint stages", func(t *testing.T) {
		stgA := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"orphan.bin": {Path: "orphan.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &entry{Stage: stgA},
			"bar.yaml": &entry{Stage: stgB},
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "foo.yaml", stageNode)

			assertGraphsEqual(expectedGraph, graph, t)
		})

		t.Run("full graph", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, false)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "foo.yaml", stageNode)
			expectedGraph.AddNode("", "orphan.bin", artifactNode)
			expectedGraph.AddNode("", "foo.bin", artifactNode)
			expectedGraph.AddEdge("foo.yaml", "orphan.bin", true, nil)
			expectedGraph.AddEdge("foo.yaml", "foo.bin", true, ownershipEdge)

			assertGraphsEqual(expectedGraph, graph, t)
		})
	})

	t.Run("connected stages", func(t *testing.T) {
		stgA := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"foo.bin": {Path: "foo.bin"},
			},
		}
		stgB := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"bar.bin": {Path: "bar.bin"},
			},
		}
		idx := Index{
			"foo.yaml": &entry{Stage: stgA},
			"bar.yaml": &entry{Stage: stgB},
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "bar.yaml", stageNode)
			expectedGraph.AddNode("", "foo.yaml", stageNode)
			expectedGraph.AddEdge("foo.yaml", "bar.yaml", true, nil)

			assertGraphsEqual(expectedGraph, graph, t)
		})

		t.Run("full graph", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, false)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "foo.yaml", stageNode)
			expectedGraph.AddNode("", "bar.yaml", stageNode)
			expectedGraph.AddNode("", "foo.bin", artifactNode)
			expectedGraph.AddNode("", "bar.bin", artifactNode)
			expectedGraph.AddEdge("foo.yaml", "foo.bin", true, ownershipEdge)
			expectedGraph.AddEdge("bar.yaml", "bar.bin", true, ownershipEdge)
			expectedGraph.AddEdge("foo.yaml", "bar.bin", true, nil)

			assertGraphsEqual(expectedGraph, graph, t)
		})
	})

	t.Run("handle skip-connections", func(t *testing.T) {
		// stgA <-- stgB <-- stgC
		//    ^---------------|
		stgA := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
		}
		idx := Index{
			"a.yaml": &entry{Stage: stgA},
			"b.yaml": &entry{Stage: stgB},
			"c.yaml": &entry{Stage: stgC},
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("c.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "a.yaml", stageNode)
			expectedGraph.AddNode("", "b.yaml", stageNode)
			expectedGraph.AddNode("", "c.yaml", stageNode)
			expectedGraph.AddEdge("c.yaml", "a.yaml", true, nil)
			expectedGraph.AddEdge("c.yaml", "b.yaml", true, nil)
			expectedGraph.AddEdge("b.yaml", "a.yaml", true, nil)

			assertGraphsEqual(expectedGraph, graph, t)
		})

		t.Run("full graph", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("c.yaml", inProgress, graph, false)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			expectedGraph.SetDir(true)
			expectedGraph.AddNode("", "a.yaml", stageNode)
			expectedGraph.AddNode("", "b.yaml", stageNode)
			expectedGraph.AddNode("", "c.yaml", stageNode)
			expectedGraph.AddNode("", "a.bin", artifactNode)
			expectedGraph.AddNode("", "b.bin", artifactNode)
			expectedGraph.AddNode("", "c.bin", artifactNode)
			expectedGraph.AddEdge("a.yaml", "a.bin", true, ownershipEdge)
			expectedGraph.AddEdge("b.yaml", "b.bin", true, ownershipEdge)
			expectedGraph.AddEdge("c.yaml", "c.bin", true, ownershipEdge)
			expectedGraph.AddEdge("c.yaml", "b.bin", true, nil)
			expectedGraph.AddEdge("c.yaml", "a.bin", true, nil)
			expectedGraph.AddEdge("b.yaml", "a.bin", true, nil)

			assertGraphsEqual(expectedGraph, graph, t)
		})
	})

	t.Run("detect cycles", func(t *testing.T) {
		// stgA <-- stgB <-- stgC --> stgD
		//    |---------------^
		stgA := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
		}
		stgB := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"a.bin": {Path: "a.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
			},
		}
		stgC := stage.Stage{
			Dependencies: map[string]*artifact.Artifact{
				"b.bin": {Path: "b.bin"},
				"d.bin": {Path: "d.bin"},
			},
			Outputs: map[string]*artifact.Artifact{
				"c.bin": {Path: "c.bin"},
			},
		}
		stgD := stage.Stage{
			Outputs: map[string]*artifact.Artifact{
				"d.bin": {Path: "d.bin"},
			},
		}
		idx := Index{
			"a.yaml": &entry{Stage: stgA},
			"b.yaml": &entry{Stage: stgB},
			"c.yaml": &entry{Stage: stgC},
			"d.yaml": &entry{Stage: stgD},
		}

		inProgress := make(map[string]bool)
		graph := gographviz.NewEscape()
		err := idx.Graph("c.yaml", inProgress, graph, true)
		if err == nil {
			t.Fatal("expected error")
		}

		expectedError := "cycle detected"
		if diff := cmp.Diff(expectedError, err.Error()); diff != "" {
			t.Fatalf("error -want +got:\n%s", diff)
		}

		expectedInProgress := map[string]bool{
			"c.yaml": true,
			"b.yaml": true,
			"a.yaml": true,
		}
		if diff := cmp.Diff(expectedInProgress, inProgress); diff != "" {
			t.Fatalf("inProgress -want +got:\n%s", diff)
		}
	})
}
