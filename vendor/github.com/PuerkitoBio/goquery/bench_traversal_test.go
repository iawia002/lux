package goquery

import (
	"testing"
)

func BenchmarkFind(b *testing.B) {
	var n int

	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = DocB().Find("dd").Length()

		} else {
			DocB().Find("dd")
		}
	}
	if n != 41 {
		b.Fatalf("want 41, got %d", n)
	}
}

func BenchmarkFindWithinSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("ul")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Find("a[class]").Length()
		} else {
			sel.Find("a[class]")
		}
	}
	if n != 39 {
		b.Fatalf("want 39, got %d", n)
	}
}

func BenchmarkFindSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("ul")
	sel2 := DocW().Find("span")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.FindSelection(sel2).Length()
		} else {
			sel.FindSelection(sel2)
		}
	}
	if n != 73 {
		b.Fatalf("want 73, got %d", n)
	}
}

func BenchmarkFindNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("ul")
	sel2 := DocW().Find("span")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.FindNodes(nodes...).Length()
		} else {
			sel.FindNodes(nodes...)
		}
	}
	if n != 73 {
		b.Fatalf("want 73, got %d", n)
	}
}

func BenchmarkContents(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-1")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Contents().Length()
		} else {
			sel.Contents()
		}
	}
	if n != 16 {
		b.Fatalf("want 16, got %d", n)
	}
}

func BenchmarkContentsFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-1")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ContentsFiltered("a[href=\"#Examples\"]").Length()
		} else {
			sel.ContentsFiltered("a[href=\"#Examples\"]")
		}
	}
	if n != 1 {
		b.Fatalf("want 1, got %d", n)
	}
}

func BenchmarkChildren(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-2")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Children().Length()
		} else {
			sel.Children()
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkChildrenFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h3")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ChildrenFiltered(".editsection").Length()
		} else {
			sel.ChildrenFiltered(".editsection")
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkParent(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Parent().Length()
		} else {
			sel.Parent()
		}
	}
	if n != 55 {
		b.Fatalf("want 55, got %d", n)
	}
}

func BenchmarkParentFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentFiltered("ul[id]").Length()
		} else {
			sel.ParentFiltered("ul[id]")
		}
	}
	if n != 4 {
		b.Fatalf("want 4, got %d", n)
	}
}

func BenchmarkParents(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("th a")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Parents().Length()
		} else {
			sel.Parents()
		}
	}
	if n != 73 {
		b.Fatalf("want 73, got %d", n)
	}
}

func BenchmarkParentsFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("th a")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsFiltered("tr").Length()
		} else {
			sel.ParentsFiltered("tr")
		}
	}
	if n != 18 {
		b.Fatalf("want 18, got %d", n)
	}
}

func BenchmarkParentsUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("th a")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsUntil("table").Length()
		} else {
			sel.ParentsUntil("table")
		}
	}
	if n != 52 {
		b.Fatalf("want 52, got %d", n)
	}
}

func BenchmarkParentsUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("th a")
	sel2 := DocW().Find("#content")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsUntilSelection(sel2).Length()
		} else {
			sel.ParentsUntilSelection(sel2)
		}
	}
	if n != 70 {
		b.Fatalf("want 70, got %d", n)
	}
}

func BenchmarkParentsUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("th a")
	sel2 := DocW().Find("#content")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsUntilNodes(nodes...).Length()
		} else {
			sel.ParentsUntilNodes(nodes...)
		}
	}
	if n != 70 {
		b.Fatalf("want 70, got %d", n)
	}
}

func BenchmarkParentsFilteredUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-1 a")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsFilteredUntil(":nth-child(1)", "ul").Length()
		} else {
			sel.ParentsFilteredUntil(":nth-child(1)", "ul")
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkParentsFilteredUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-1 a")
	sel2 := DocW().Find("ul")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsFilteredUntilSelection(":nth-child(1)", sel2).Length()
		} else {
			sel.ParentsFilteredUntilSelection(":nth-child(1)", sel2)
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkParentsFilteredUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find(".toclevel-1 a")
	sel2 := DocW().Find("ul")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ParentsFilteredUntilNodes(":nth-child(1)", nodes...).Length()
		} else {
			sel.ParentsFilteredUntilNodes(":nth-child(1)", nodes...)
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkSiblings(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("ul li:nth-child(1)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Siblings().Length()
		} else {
			sel.Siblings()
		}
	}
	if n != 293 {
		b.Fatalf("want 293, got %d", n)
	}
}

func BenchmarkSiblingsFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("ul li:nth-child(1)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.SiblingsFiltered("[class]").Length()
		} else {
			sel.SiblingsFiltered("[class]")
		}
	}
	if n != 46 {
		b.Fatalf("want 46, got %d", n)
	}
}

func BenchmarkNext(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(1)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Next().Length()
		} else {
			sel.Next()
		}
	}
	if n != 49 {
		b.Fatalf("want 49, got %d", n)
	}
}

func BenchmarkNextFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(1)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextFiltered("[class]").Length()
		} else {
			sel.NextFiltered("[class]")
		}
	}
	if n != 6 {
		b.Fatalf("want 6, got %d", n)
	}
}

func BenchmarkNextAll(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(3)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextAll().Length()
		} else {
			sel.NextAll()
		}
	}
	if n != 234 {
		b.Fatalf("want 234, got %d", n)
	}
}

func BenchmarkNextAllFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(3)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextAllFiltered("[class]").Length()
		} else {
			sel.NextAllFiltered("[class]")
		}
	}
	if n != 33 {
		b.Fatalf("want 33, got %d", n)
	}
}

func BenchmarkPrev(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:last-child")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Prev().Length()
		} else {
			sel.Prev()
		}
	}
	if n != 49 {
		b.Fatalf("want 49, got %d", n)
	}
}

func BenchmarkPrevFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:last-child")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevFiltered("[class]").Length()
		} else {
			sel.PrevFiltered("[class]")
		}
	}
	// There is one more Prev li with a class, compared to Next li with a class
	// (confirmed by looking at the HTML, this is ok)
	if n != 7 {
		b.Fatalf("want 7, got %d", n)
	}
}

func BenchmarkPrevAll(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(4)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevAll().Length()
		} else {
			sel.PrevAll()
		}
	}
	if n != 78 {
		b.Fatalf("want 78, got %d", n)
	}
}

func BenchmarkPrevAllFiltered(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:nth-child(4)")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevAllFiltered("[class]").Length()
		} else {
			sel.PrevAllFiltered("[class]")
		}
	}
	if n != 6 {
		b.Fatalf("want 6, got %d", n)
	}
}

func BenchmarkNextUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:first-child")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextUntil(":nth-child(4)").Length()
		} else {
			sel.NextUntil(":nth-child(4)")
		}
	}
	if n != 84 {
		b.Fatalf("want 84, got %d", n)
	}
}

func BenchmarkNextUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("ul")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextUntilSelection(sel2).Length()
		} else {
			sel.NextUntilSelection(sel2)
		}
	}
	if n != 42 {
		b.Fatalf("want 42, got %d", n)
	}
}

func BenchmarkNextUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("p")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextUntilNodes(nodes...).Length()
		} else {
			sel.NextUntilNodes(nodes...)
		}
	}
	if n != 12 {
		b.Fatalf("want 12, got %d", n)
	}
}

func BenchmarkPrevUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("li:last-child")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevUntil(":nth-child(4)").Length()
		} else {
			sel.PrevUntil(":nth-child(4)")
		}
	}
	if n != 238 {
		b.Fatalf("want 238, got %d", n)
	}
}

func BenchmarkPrevUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("ul")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevUntilSelection(sel2).Length()
		} else {
			sel.PrevUntilSelection(sel2)
		}
	}
	if n != 49 {
		b.Fatalf("want 49, got %d", n)
	}
}

func BenchmarkPrevUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("p")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevUntilNodes(nodes...).Length()
		} else {
			sel.PrevUntilNodes(nodes...)
		}
	}
	if n != 11 {
		b.Fatalf("want 11, got %d", n)
	}
}

func BenchmarkNextFilteredUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextFilteredUntil("p", "div").Length()
		} else {
			sel.NextFilteredUntil("p", "div")
		}
	}
	if n != 22 {
		b.Fatalf("want 22, got %d", n)
	}
}

func BenchmarkNextFilteredUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("div")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextFilteredUntilSelection("p", sel2).Length()
		} else {
			sel.NextFilteredUntilSelection("p", sel2)
		}
	}
	if n != 22 {
		b.Fatalf("want 22, got %d", n)
	}
}

func BenchmarkNextFilteredUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("div")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.NextFilteredUntilNodes("p", nodes...).Length()
		} else {
			sel.NextFilteredUntilNodes("p", nodes...)
		}
	}
	if n != 22 {
		b.Fatalf("want 22, got %d", n)
	}
}

func BenchmarkPrevFilteredUntil(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevFilteredUntil("p", "div").Length()
		} else {
			sel.PrevFilteredUntil("p", "div")
		}
	}
	if n != 20 {
		b.Fatalf("want 20, got %d", n)
	}
}

func BenchmarkPrevFilteredUntilSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("div")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevFilteredUntilSelection("p", sel2).Length()
		} else {
			sel.PrevFilteredUntilSelection("p", sel2)
		}
	}
	if n != 20 {
		b.Fatalf("want 20, got %d", n)
	}
}

func BenchmarkPrevFilteredUntilNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := DocW().Find("h2")
	sel2 := DocW().Find("div")
	nodes := sel2.Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.PrevFilteredUntilNodes("p", nodes...).Length()
		} else {
			sel.PrevFilteredUntilNodes("p", nodes...)
		}
	}
	if n != 20 {
		b.Fatalf("want 20, got %d", n)
	}
}

func BenchmarkClosest(b *testing.B) {
	var n int

	b.StopTimer()
	sel := Doc().Find(".container-fluid")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.Closest(".pvk-content").Length()
		} else {
			sel.Closest(".pvk-content")
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkClosestSelection(b *testing.B) {
	var n int

	b.StopTimer()
	sel := Doc().Find(".container-fluid")
	sel2 := Doc().Find(".pvk-content")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ClosestSelection(sel2).Length()
		} else {
			sel.ClosestSelection(sel2)
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}

func BenchmarkClosestNodes(b *testing.B) {
	var n int

	b.StopTimer()
	sel := Doc().Find(".container-fluid")
	nodes := Doc().Find(".pvk-content").Nodes
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if n == 0 {
			n = sel.ClosestNodes(nodes...).Length()
		} else {
			sel.ClosestNodes(nodes...)
		}
	}
	if n != 2 {
		b.Fatalf("want 2, got %d", n)
	}
}
