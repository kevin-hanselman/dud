package index

import (
	"testing"

	"github.com/awalterschulze/gographviz"
	"github.com/google/go-cmp/cmp"
	"github.com/kevin-hanselman/dud/src/artifact"
	"github.com/kevin-hanselman/dud/src/stage"
)

func check(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func assertGraphsEqual(
	graphWant *gographviz.Escape,
	graphGot *gographviz.Escape,
	t *testing.T,
) {
	if graphWant.Directed != graphGot.Directed {
		t.Fatalf("graph.Directed = %v", graphGot.Directed)
	}
	if graphWant.Strict != graphGot.Strict {
		t.Fatalf("graph.Strict = %v", graphGot.Strict)
	}
	if diff := cmp.Diff(graphWant.Attrs, graphGot.Attrs); diff != "" {
		t.Fatalf("graph.Attrs -want +got:\n%s", diff)
	}
	if diff := cmp.Diff(graphWant.Nodes.Sorted(), graphGot.Nodes.Sorted()); diff != "" {
		t.Fatalf("graph.Nodes -want +got:\n%s", diff)
	}
	if diff := cmp.Diff(graphWant.Edges.Sorted(), graphGot.Edges.Sorted()); diff != "" {
		t.Fatalf("graph.Edges -want +got:\n%s", diff)
	}
	// Ignore styling on subgraphs; it's too cumbersome and brittle to test.
	subGot := graphGot.SubGraphs.Sorted()
	for _, subgraph := range subGot {
		subgraph.Attrs = gographviz.Attrs{}
	}
	if diff := cmp.Diff(graphWant.SubGraphs.Sorted(), subGot); diff != "" {
		t.Fatalf("graph.SubGraphs -want +got:\n%s", diff)
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
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			check(t, expectedGraph.AddNode("", "foo.yaml", nil))

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
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			// See stageSubGraphName for explanation of "cluster" prefix.
			check(t, expectedGraph.AddSubGraph("", "cluster_foo.yaml", nil))
			check(t, expectedGraph.AddNode("cluster_foo.yaml", "foo.yaml", hiddenAttr))
			check(t, expectedGraph.AddNode("", "orphan.bin", nil))
			check(t, expectedGraph.AddNode("cluster_foo.yaml", "foo.bin", nil))
			check(t, expectedGraph.AddEdge("foo.yaml", "orphan.bin", true, map[string]string{"ltail": "cluster_foo.yaml"}))

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
			"foo.yaml": &stgA,
			"bar.yaml": &stgB,
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("foo.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			check(t, expectedGraph.AddNode("", "bar.yaml", nil))
			check(t, expectedGraph.AddNode("", "foo.yaml", nil))
			check(t, expectedGraph.AddEdge("foo.yaml", "bar.yaml", true, nil))

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
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			check(t, expectedGraph.AddSubGraph("", "cluster_foo.yaml", nil))
			check(t, expectedGraph.AddNode("cluster_foo.yaml", "foo.yaml", hiddenAttr))
			check(t, expectedGraph.AddSubGraph("", "cluster_bar.yaml", nil))
			check(t, expectedGraph.AddNode("cluster_bar.yaml", "bar.yaml", hiddenAttr))
			check(t, expectedGraph.AddNode("cluster_foo.yaml", "foo.bin", nil))
			check(t, expectedGraph.AddNode("cluster_bar.yaml", "bar.bin", nil))
			check(t, expectedGraph.AddEdge("foo.yaml", "bar.bin", true, map[string]string{"ltail": "cluster_foo.yaml"}))

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
			"a.yaml": &stgA,
			"b.yaml": &stgB,
			"c.yaml": &stgC,
		}

		t.Run("only stages", func(t *testing.T) {
			inProgress := make(map[string]bool)
			graph := gographviz.NewEscape()
			err := idx.Graph("c.yaml", inProgress, graph, true)
			if err != nil {
				t.Fatal(err)
			}

			expectedGraph := gographviz.NewEscape()
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			check(t, expectedGraph.AddNode("", "a.yaml", nil))
			check(t, expectedGraph.AddNode("", "b.yaml", nil))
			check(t, expectedGraph.AddNode("", "c.yaml", nil))
			check(t, expectedGraph.AddEdge("c.yaml", "a.yaml", true, nil))
			check(t, expectedGraph.AddEdge("c.yaml", "b.yaml", true, nil))
			check(t, expectedGraph.AddEdge("b.yaml", "a.yaml", true, nil))

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
			check(t, expectedGraph.SetDir(true))
			check(t, expectedGraph.SetStrict(true))
			check(t, expectedGraph.Attrs.Add("compound", "true"))
			check(t, expectedGraph.Attrs.Add("rankdir", "LR"))
			check(t, expectedGraph.AddSubGraph("", "cluster_a.yaml", nil))
			check(t, expectedGraph.AddSubGraph("", "cluster_b.yaml", nil))
			check(t, expectedGraph.AddSubGraph("", "cluster_c.yaml", nil))
			check(t, expectedGraph.AddNode("cluster_a.yaml", "a.yaml", hiddenAttr))
			check(t, expectedGraph.AddNode("cluster_b.yaml", "b.yaml", hiddenAttr))
			check(t, expectedGraph.AddNode("cluster_c.yaml", "c.yaml", hiddenAttr))
			check(t, expectedGraph.AddNode("cluster_a.yaml", "a.bin", nil))
			check(t, expectedGraph.AddNode("cluster_b.yaml", "b.bin", nil))
			check(t, expectedGraph.AddNode("cluster_c.yaml", "c.bin", nil))
			check(t, expectedGraph.AddEdge("c.yaml", "b.bin", true, map[string]string{"ltail": "cluster_c.yaml"}))
			check(t, expectedGraph.AddEdge("c.yaml", "a.bin", true, map[string]string{"ltail": "cluster_c.yaml"}))
			check(t, expectedGraph.AddEdge("b.yaml", "a.bin", true, map[string]string{"ltail": "cluster_b.yaml"}))

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
			"a.yaml": &stgA,
			"b.yaml": &stgB,
			"c.yaml": &stgC,
			"d.yaml": &stgD,
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
