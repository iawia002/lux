package goquery

import (
	"testing"
)

func TestAdd(t *testing.T) {
	sel := Doc().Find("div.row-fluid").Add("a")
	assertLength(t, sel.Nodes, 19)
}

func TestAddInvalid(t *testing.T) {
	sel1 := Doc().Find("div.row-fluid")
	sel2 := sel1.Add("")
	assertLength(t, sel1.Nodes, 9)
	assertLength(t, sel2.Nodes, 9)
	if sel1 == sel2 {
		t.Errorf("selections should not be the same")
	}
}

func TestAddRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Add("a").End()
	assertEqual(t, sel, sel2)
}

func TestAddSelection(t *testing.T) {
	sel := Doc().Find("div.row-fluid")
	sel2 := Doc().Find("a")
	sel = sel.AddSelection(sel2)
	assertLength(t, sel.Nodes, 19)
}

func TestAddSelectionNil(t *testing.T) {
	sel := Doc().Find("div.row-fluid")
	assertLength(t, sel.Nodes, 9)

	sel = sel.AddSelection(nil)
	assertLength(t, sel.Nodes, 9)
}

func TestAddSelectionRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a")
	sel2 = sel.AddSelection(sel2).End()
	assertEqual(t, sel, sel2)
}

func TestAddNodes(t *testing.T) {
	sel := Doc().Find("div.pvk-gutter")
	sel2 := Doc().Find(".pvk-content")
	sel = sel.AddNodes(sel2.Nodes...)
	assertLength(t, sel.Nodes, 9)
}

func TestAddNodesNone(t *testing.T) {
	sel := Doc().Find("div.pvk-gutter").AddNodes()
	assertLength(t, sel.Nodes, 6)
}

func TestAddNodesRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a")
	sel2 = sel.AddNodes(sel2.Nodes...).End()
	assertEqual(t, sel, sel2)
}

func TestAddNodesBig(t *testing.T) {
	doc := DocW()
	sel := doc.Find("li")
	assertLength(t, sel.Nodes, 373)
	sel2 := doc.Find("xyz")
	assertLength(t, sel2.Nodes, 0)

	nodes := sel.Nodes
	sel2 = sel2.AddNodes(nodes...)
	assertLength(t, sel2.Nodes, 373)
	nodes2 := append(nodes, nodes...)
	sel2 = sel2.End().AddNodes(nodes2...)
	assertLength(t, sel2.Nodes, 373)
	nodes3 := append(nodes2, nodes...)
	sel2 = sel2.End().AddNodes(nodes3...)
	assertLength(t, sel2.Nodes, 373)
}

func TestAndSelf(t *testing.T) {
	sel := Doc().Find(".span12").Last().AndSelf()
	assertLength(t, sel.Nodes, 2)
}

func TestAndSelfRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a").AndSelf().End().End()
	assertEqual(t, sel, sel2)
}

func TestAddBack(t *testing.T) {
	sel := Doc().Find(".span12").Last().AddBack()
	assertLength(t, sel.Nodes, 2)
}

func TestAddBackRollback(t *testing.T) {
	sel := Doc().Find(".pvk-content")
	sel2 := sel.Find("a").AddBack().End().End()
	assertEqual(t, sel, sel2)
}

func TestAddBackFiltered(t *testing.T) {
	sel := Doc().Find(".span12, .footer").Find("h1").AddBackFiltered(".footer")
	assertLength(t, sel.Nodes, 2)
}

func TestAddBackFilteredRollback(t *testing.T) {
	sel := Doc().Find(".span12, .footer")
	sel2 := sel.Find("h1").AddBackFiltered(".footer").End().End()
	assertEqual(t, sel, sel2)
}
