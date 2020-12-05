package index

import (
	"github.com/awalterschulze/gographviz"
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"

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
	if diff := cmp.Diff(graphWant.SubGraphs.Sorted(), graphGot.SubGraphs.Sorted()); diff != "" {
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
			expectedGraph.AddNode("", "foo.yaml", nil)

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
			// See stageSubGraphName for explanation of "cluster" prefix.
			expectedGraph.AddSubGraph("", "cluster_foo.yaml", map[string]string{"label": "foo.yaml"})
			expectedGraph.AddNode("", "orphan.bin", nil)
			expectedGraph.AddNode("cluster_foo.yaml", "foo.bin", nil)
			expectedGraph.AddEdge("cluster_foo.yaml", "orphan.bin", true, nil)

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
			expectedGraph.AddNode("", "bar.yaml", nil)
			expectedGraph.AddNode("", "foo.yaml", nil)
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
			expectedGraph.AddSubGraph("", "cluster_foo.yaml", map[string]string{"label": "foo.yaml"})
			expectedGraph.AddSubGraph("", "cluster_bar.yaml", map[string]string{"label": "bar.yaml"})
			expectedGraph.AddNode("cluster_foo.yaml", "foo.bin", nil)
			expectedGraph.AddNode("cluster_bar.yaml", "bar.bin", nil)
			expectedGraph.AddEdge("cluster_foo.yaml", "bar.bin", true, nil)

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
			expectedGraph.AddNode("", "a.yaml", nil)
			expectedGraph.AddNode("", "b.yaml", nil)
			expectedGraph.AddNode("", "c.yaml", nil)
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
			expectedGraph.AddSubGraph("", "cluster_a.yaml", map[string]string{"label": "a.yaml"})
			expectedGraph.AddSubGraph("", "cluster_b.yaml", map[string]string{"label": "b.yaml"})
			expectedGraph.AddSubGraph("", "cluster_c.yaml", map[string]string{"label": "c.yaml"})
			expectedGraph.AddNode("cluster_a.yaml", "a.bin", nil)
			expectedGraph.AddNode("cluster_b.yaml", "b.bin", nil)
			expectedGraph.AddNode("cluster_c.yaml", "c.bin", nil)
			expectedGraph.AddEdge("cluster_c.yaml", "b.bin", true, nil)
			expectedGraph.AddEdge("cluster_c.yaml", "a.bin", true, nil)
			expectedGraph.AddEdge("cluster_b.yaml", "a.bin", true, nil)

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

		onlyStages := true

		test := func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("c.yaml", inProgress, graph, onlyStages)
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
		}
		t.Run("only stages", test)

		onlyStages = false
		t.Run("full graph", test)
	})
}
