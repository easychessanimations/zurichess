package main

import "testing"

const (
	// These constants should change only when search/evaluation is changed.
	// Non-functional changes should not change the number of nodes.
	shallowDepth = 4
	shallowNodes = 4053846
	deepDepth    = 5
	deepNodes    = 13041207
)

func TestShallow(t *testing.T) {
	nodes, _ := evalAll(shallowDepth)
	if shallowNodes != nodes {
		t.Fatalf("at depth %d expected %d nodes, got %d", shallowDepth, shallowNodes, nodes)
	}

}

func TestDeep(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	nodes, _ := evalAll(deepDepth)
	if deepNodes != nodes {
		t.Fatalf("at depth %d expected %d nodes, got %d", deepDepth, deepNodes, nodes)
	}
}
