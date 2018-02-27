package cascadia

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

type selectorTest struct {
	HTML, selector string
	results        []string
}

func nodeString(n *html.Node) string {
	buf := bytes.NewBufferString("")
	html.Render(buf, n)
	return buf.String()
}

var selectorTests = []selectorTest{
	{
		`<body><address>This address...</address></body>`,
		"address",
		[]string{
			"<address>This address...</address>",
		},
	},
	{
		`<!-- comment --><html><head></head><body>text</body></html>`,
		"*",
		[]string{
			"<html><head></head><body>text</body></html>",
			"<head></head>",
			"<body>text</body>",
		},
	},
	{
		`<html><head></head><body></body></html>`,
		"*",
		[]string{
			"<html><head></head><body></body></html>",
			"<head></head>",
			"<body></body>",
		},
	},
	{
		`<p id="foo"><p id="bar">`,
		"#foo",
		[]string{
			`<p id="foo"></p>`,
		},
	},
	{
		`<ul><li id="t1"><p id="t1">`,
		"li#t1",
		[]string{
			`<li id="t1"><p id="t1"></p></li>`,
		},
	},
	{
		`<ol><li id="t4"><li id="t44">`,
		"*#t4",
		[]string{
			`<li id="t4"></li>`,
		},
	},
	{
		`<ul><li class="t1"><li class="t2">`,
		".t1",
		[]string{
			`<li class="t1"></li>`,
		},
	},
	{
		`<p class="t1 t2">`,
		"p.t1",
		[]string{
			`<p class="t1 t2"></p>`,
		},
	},
	{
		`<div class="test">`,
		"div.teST",
		[]string{},
	},
	{
		`<p class="t1 t2">`,
		".t1.fail",
		[]string{},
	},
	{
		`<p class="t1 t2">`,
		"p.t1.t2",
		[]string{
			`<p class="t1 t2"></p>`,
		},
	},
	{
		`<p><p title="title">`,
		"p[title]",
		[]string{
			`<p title="title"></p>`,
		},
	},
	{
		`<address><address title="foo"><address title="bar">`,
		`address[title="foo"]`,
		[]string{
			`<address title="foo"><address title="bar"></address></address>`,
		},
	},
	{
		`<address><address title="foo"><address title="bar">`,
		`address[title!="foo"]`,
		[]string{
			`<address><address title="foo"><address title="bar"></address></address></address>`,
			`<address title="bar"></address>`,
		},
	},
	{
		`<p title="tot foo bar">`,
		`[    	title        ~=       foo    ]`,
		[]string{
			`<p title="tot foo bar"></p>`,
		},
	},
	{
		`<p title="hello world">`,
		`[title~="hello world"]`,
		[]string{},
	},
	{
		`<p lang="en"><p lang="en-gb"><p lang="enough"><p lang="fr-en">`,
		`[lang|="en"]`,
		[]string{
			`<p lang="en"></p>`,
			`<p lang="en-gb"></p>`,
		},
	},
	{
		`<p title="foobar"><p title="barfoo">`,
		`[title^="foo"]`,
		[]string{
			`<p title="foobar"></p>`,
		},
	},
	{
		`<p title="foobar"><p title="barfoo">`,
		`[title$="bar"]`,
		[]string{
			`<p title="foobar"></p>`,
		},
	},
	{
		`<p title="foobarufoo">`,
		`[title*="bar"]`,
		[]string{
			`<p title="foobarufoo"></p>`,
		},
	},
	{
		`<p class=" ">This text should be green.</p><p>This text should be green.</p>`,
		`p[class$=" "]`,
		[]string{},
	},
	{
		`<p class="">This text should be green.</p><p>This text should be green.</p>`,
		`p[class$=""]`,
		[]string{},
	},
	{
		`<p class=" ">This text should be green.</p><p>This text should be green.</p>`,
		`p[class^=" "]`,
		[]string{},
	},
	{
		`<p class="">This text should be green.</p><p>This text should be green.</p>`,
		`p[class^=""]`,
		[]string{},
	},
	{
		`<p class=" ">This text should be green.</p><p>This text should be green.</p>`,
		`p[class*=" "]`,
		[]string{},
	},
	{
		`<p class="">This text should be green.</p><p>This text should be green.</p>`,
		`p[class*=""]`,
		[]string{},
	},
	{
		`<input type="radio" name="Sex" value="F"/>`,
		`input[name=Sex][value=F]`,
		[]string{
			`<input type="radio" name="Sex" value="F"/>`,
		},
	},
	{
		`<table border="0" cellpadding="0" cellspacing="0" style="table-layout: fixed; width: 100%; border: 0 dashed; border-color: #FFFFFF"><tr style="height:64px">aaa</tr></table>`,
		`table[border="0"][cellpadding="0"][cellspacing="0"]`,
		[]string{
			`<table border="0" cellpadding="0" cellspacing="0" style="table-layout: fixed; width: 100%; border: 0 dashed; border-color: #FFFFFF"><tbody><tr style="height:64px"></tr></tbody></table>`,
		},
	},
	{
		`<p class="t1 t2">`,
		".t1:not(.t2)",
		[]string{},
	},
	{
		`<div class="t3">`,
		`div:not(.t1)`,
		[]string{
			`<div class="t3"></div>`,
		},
	},
	{
		`<div><div class="t2"><div class="t3">`,
		`div:not([class="t2"])`,
		[]string{
			`<div><div class="t2"><div class="t3"></div></div></div>`,
			`<div class="t3"></div>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(odd)`,
		[]string{
			`<li id="1"></li>`,
			`<li id="3"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(even)`,
		[]string{
			`<li id="2"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(-n+2)`,
		[]string{
			`<li id="1"></li>`,
			`<li id="2"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3></ol>`,
		`li:nth-child(3n+1)`,
		[]string{
			`<li id="1"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(odd)`,
		[]string{
			`<li id="2"></li>`,
			`<li id="4"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(even)`,
		[]string{
			`<li id="1"></li>`,
			`<li id="3"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(-n+2)`,
		[]string{
			`<li id="3"></li>`,
			`<li id="4"></li>`,
		},
	},
	{
		`<ol><li id=1><li id=2><li id=3><li id=4></ol>`,
		`li:nth-last-child(3n+1)`,
		[]string{
			`<li id="1"></li>`,
			`<li id="4"></li>`,
		},
	},
	{
		`<p>some text <span id="1">and a span</span><span id="2"> and another</span></p>`,
		`span:first-child`,
		[]string{
			`<span id="1">and a span</span>`,
		},
	},
	{
		`<span>a span</span> and some text`,
		`span:last-child`,
		[]string{
			`<span>a span</span>`,
		},
	},
	{
		`<address></address><p id=1><p id=2>`,
		`p:nth-of-type(2)`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:nth-last-of-type(2)`,
		[]string{
			`<p id="1"></p>`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:last-of-type`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<address></address><p id=1><p id=2></p><a>`,
		`p:first-of-type`,
		[]string{
			`<p id="1"></p>`,
		},
	},
	{
		`<div><p id="1"></p><a></a></div><div><p id="2"></p></div>`,
		`p:only-child`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<div><p id="1"></p><a></a></div><div><p id="2"></p><p id="3"></p></div>`,
		`p:only-of-type`,
		[]string{
			`<p id="1"></p>`,
		},
	},
	{
		`<p id="1"><!-- --><p id="2">Hello<p id="3"><span>`,
		`:empty`,
		[]string{
			`<head></head>`,
			`<p id="1"><!-- --></p>`,
			`<span></span>`,
		},
	},
	{
		`<div><p id="1"><table><tr><td><p id="2"></table></div><p id="3">`,
		`div p`,
		[]string{
			`<p id="1"><table><tbody><tr><td><p id="2"></p></td></tr></tbody></table></p>`,
			`<p id="2"></p>`,
		},
	},
	{
		`<div><p id="1"><table><tr><td><p id="2"></table></div><p id="3">`,
		`div table p`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<div><p id="1"><div><p id="2"></div><table><tr><td><p id="3"></table></div>`,
		`div > p`,
		[]string{
			`<p id="1"></p>`,
			`<p id="2"></p>`,
		},
	},
	{
		`<p id="1"><p id="2"></p><address></address><p id="3">`,
		`p ~ p`,
		[]string{
			`<p id="2"></p>`,
			`<p id="3"></p>`,
		},
	},
	{
		`<p id="1"></p>
		 <!--comment-->
		 <p id="2"></p><address></address><p id="3">`,
		`p + p`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<ul><li></li><li></li></ul><p>`,
		`li, p`,
		[]string{
			"<li></li>",
			"<li></li>",
			"<p></p>",
		},
	},
	{
		`<p id="1"><p id="2"></p><address></address><p id="3">`,
		`p +/*This is a comment*/ p`,
		[]string{
			`<p id="2"></p>`,
		},
	},
	{
		`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		`p:contains("that wraps")`,
		[]string{
			`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		},
	},
	{
		`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		`p:containsOwn("that wraps")`,
		[]string{},
	},
	{
		`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		`:containsOwn("inner")`,
		[]string{
			`<span>wraps inner text</span>`,
		},
	},
	{
		`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		`p:containsOwn("block")`,
		[]string{
			`<p>Text block that <span>wraps inner text</span> and continues</p>`,
		},
	},
	{
		`<div id="d1"><p id="p1"><span>text content</span></p></div><div id="d2"/>`,
		`div:has(#p1)`,
		[]string{
			`<div id="d1"><p id="p1"><span>text content</span></p></div>`,
		},
	},
	{
		`<div id="d1"><p id="p1"><span>contents 1</span></p></div>
		<div id="d2"><p>contents <em>2</em></p></div>`,
		`div:has(:containsOwn("2"))`,
		[]string{
			`<div id="d2"><p>contents <em>2</em></p></div>`,
		},
	},
	{
		`<body><div id="d1"><p id="p1"><span>contents 1</span></p></div>
		<div id="d2"><p id="p2">contents <em>2</em></p></div></body>`,
		`body :has(:containsOwn("2"))`,
		[]string{
			`<div id="d2"><p id="p2">contents <em>2</em></p></div>`,
			`<p id="p2">contents <em>2</em></p>`,
		},
	},
	{
		`<body><div id="d1"><p id="p1"><span>contents 1</span></p></div>
		<div id="d2"><p id="p2">contents <em>2</em></p></div></body>`,
		`body :haschild(:containsOwn("2"))`,
		[]string{
			`<p id="p2">contents <em>2</em></p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches([\d])`,
		[]string{
			`<p id="p1">0123456789</p>`,
			`<p id="p3">0123ABCD</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches([a-z])`,
		[]string{
			`<p id="p2">abcdef</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches([a-zA-Z])`,
		[]string{
			`<p id="p2">abcdef</p>`,
			`<p id="p3">0123ABCD</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches([^\d])`,
		[]string{
			`<p id="p2">abcdef</p>`,
			`<p id="p3">0123ABCD</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches(^(0|a))`,
		[]string{
			`<p id="p1">0123456789</p>`,
			`<p id="p2">abcdef</p>`,
			`<p id="p3">0123ABCD</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:matches(^\d+$)`,
		[]string{
			`<p id="p1">0123456789</p>`,
		},
	},
	{
		`<p id="p1">0123456789</p><p id="p2">abcdef</p><p id="p3">0123ABCD</p>`,
		`p:not(:matches(^\d+$))`,
		[]string{
			`<p id="p2">abcdef</p>`,
			`<p id="p3">0123ABCD</p>`,
		},
	},
	{
		`<div><p id="p1">01234<em>567</em>89</p><div>`,
		`div :matchesOwn(^\d+$)`,
		[]string{
			`<p id="p1">01234<em>567</em>89</p>`,
			`<em>567</em>`,
		},
	},
	{
		`<ul>
			<li><a id="a1" href="http://www.google.com/finance"></a>
			<li><a id="a2" href="http://finance.yahoo.com/"></a>
			<li><a id="a2" href="http://finance.untrusted.com/"/>
			<li><a id="a3" href="https://www.google.com/news"/>
			<li><a id="a4" href="http://news.yahoo.com"/>
		</ul>`,
		`[href#=(fina)]:not([href#=(\/\/[^\/]+untrusted)])`,
		[]string{
			`<a id="a1" href="http://www.google.com/finance"></a>`,
			`<a id="a2" href="http://finance.yahoo.com/"></a>`,
		},
	},
	{
		`<ul>
			<li><a id="a1" href="http://www.google.com/finance"/>
			<li><a id="a2" href="http://finance.yahoo.com/"/>
			<li><a id="a3" href="https://www.google.com/news"></a>
			<li><a id="a4" href="http://news.yahoo.com"/>
		</ul>`,
		`[href#=(^https:\/\/[^\/]*\/?news)]`,
		[]string{
			`<a id="a3" href="https://www.google.com/news"></a>`,
		},
	},
	{
		`<form>
			<label>Username <input type="text" name="username" /></label>
			<label>Password <input type="password" name="password" /></label>
			<label>Country
				<select name="country">
					<option value="ca">Canada</option>
					<option value="us">United States</option>
				</select>
			</label>
			<label>Bio <textarea name="bio"></textarea></label>
			<button>Sign up</button>
		</form>`,
		`:input`,
		[]string{
			`<input type="text" name="username"/>`,
			`<input type="password" name="password"/>`,
			`<select name="country">
					<option value="ca">Canada</option>
					<option value="us">United States</option>
				</select>`,
			`<textarea name="bio"></textarea>`,
			`<button>Sign up</button>`,
		},
	},
	{
		`<html><head></head><body></body></html>`,
		":root",
		[]string{
			"<html><head></head><body></body></html>",
		},
	},
	{
		`<html><head></head><body></body></html>`,
		"*:root",
		[]string{
			"<html><head></head><body></body></html>",
		},
	},
	{
		`<html><head></head><body></body></html>`,
		"*:root:first-child",
		[]string{},
	},
	{
		`<html><head></head><body></body></html>`,
		"*:root:nth-child(1)",
		[]string{},
	},
	{
		`<html><head></head><body><a href="http://www.foo.com"></a></body></html>`,
		"a:not(:root)",
		[]string{
			`<a href="http://www.foo.com"></a>`,
		},
	},
}

func TestSelectors(t *testing.T) {
	for _, test := range selectorTests {
		s, err := Compile(test.selector)
		if err != nil {
			t.Errorf("error compiling %q: %s", test.selector, err)
			continue
		}

		doc, err := html.Parse(strings.NewReader(test.HTML))
		if err != nil {
			t.Errorf("error parsing %q: %s", test.HTML, err)
			continue
		}

		matches := s.MatchAll(doc)
		if len(matches) != len(test.results) {
			t.Errorf("selector %s wanted %d elements, got %d instead", test.selector, len(test.results), len(matches))
			continue
		}

		for i, m := range matches {
			got := nodeString(m)
			if got != test.results[i] {
				t.Errorf("selector %s wanted %s, got %s instead", test.selector, test.results[i], got)
			}
		}

		firstMatch := s.MatchFirst(doc)
		if len(test.results) == 0 {
			if firstMatch != nil {
				t.Errorf("MatchFirst: selector %s want nil, got %s", test.selector, nodeString(firstMatch))
			}
		} else {
			got := nodeString(firstMatch)
			if got != test.results[0] {
				t.Errorf("MatchFirst: selector %s want %s, got %s", test.selector, test.results[0], got)
			}
		}
	}
}
