package goquery

import (
	"testing"
)

func BenchmarkIs(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("li")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.Is(".toclevel-2")
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkIsPositional(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("li")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.Is("li:nth-child(2)")
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkIsFunction(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find(".toclevel-1")
	f := func(i int, s *Selection) bool {
		return i == 8
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.IsFunction(f)
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkIsSelection(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("li")
	sel2 := DocW().Find(".toclevel-2")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.IsSelection(sel2)
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkIsNodes(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("li")
	sel2 := DocW().Find(".toclevel-2")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.IsNodes(nodes...)
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkHasClass(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("span")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.HasClass("official")
	}
	if !y {
		b.Fatal("want true")
	}
}

func BenchmarkContains(b *testing.B) {
	var y bool

	b.StopTimer()
	sel := DocW().Find("span.url")
	sel2 := DocW().Find("a[rel=\"nofollow\"]")
	node := sel2.Nodes[0]
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		y = sel.Contains(node)
	}
	if !y {
		b.Fatal("want true")
	}
}
