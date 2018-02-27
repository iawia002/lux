package goquery

import (
	"testing"
)

func BenchmarkAdd(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocB().Find("dd")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Add("h2[title]").Length()
		} else {
			sel.Add("h2[title]")
		}
	}
	if n != 43 {
		b.Fatalf("want 43, got %d", n)
	}
}

func BenchmarkAddSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocB().Find("dd")
	sel2 := DocB().Find("h2[title]")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.AddSelection(sel2).Length()
		} else {
			sel.AddSelection(sel2)
		}
	}
	if n != 43 {
		b.Fatalf("want 43, got %d", n)
	}
}

func BenchmarkAddNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocB().Find("dd")
	sel2 := DocB().Find("h2[title]")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.AddNodes(nodes...).Length()
		} else {
			sel.AddNodes(nodes...)
		}
	}
	if n != 43 {
		b.Fatalf("want 43, got %d", n)
	}
}

func BenchmarkAddNodesBig(b *testing.B) {
	var n int

	doc := DocW()
	sel := doc.Find("li")
	// make nodes > 1000
	nodes := sel.Nodes
	nodes = append(nodes, nodes...)
	nodes = append(nodes, nodes...)
	sel = doc.Find("xyz")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.AddNodes(nodes...).Length()
		} else {
			sel.AddNodes(nodes...)
		}
	}
	if n != 373 {
		b.Fatalf("want 373, got %d", n)
	}
}

func BenchmarkAndSelf(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocB().Find("dd").Parent()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.AndSelf().Length()
		} else {
			sel.AndSelf()
		}
	}
	if n != 44 {
		b.Fatalf("want 44, got %d", n)
	}
}
