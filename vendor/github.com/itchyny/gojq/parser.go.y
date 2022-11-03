%{
package gojq

// Parse parses a query.
func Parse(src string) (*Query, error) {
	l := newLexer(src)
	if yyParse(l) > 0 {
		return nil, l.err
	}
	return l.result, nil
}

func reverseFuncDef(xs []*FuncDef) []*FuncDef {
	for i, j := 0, len(xs)-1; i < j; i, j = i+1, j-1 {
		xs[i], xs[j] = xs[j], xs[i]
	}
	return xs
}

func prependFuncDef(xs []*FuncDef, x *FuncDef) []*FuncDef {
	xs = append(xs, nil)
	copy(xs[1:], xs)
	xs[0] = x
	return xs
}
%}

%union {
  value    interface{}
  token    string
  operator Operator
}

%type<value> program moduleheader programbody imports import metaopt funcdefs funcdef funcdefargs query
%type<value> bindpatterns pattern arraypatterns objectpatterns objectpattern
%type<value> term string stringparts suffix args ifelifs ifelse trycatch
%type<value> objectkeyvals objectkeyval objectval
%type<value> constterm constobject constobjectkeyvals constobjectkeyval constarray constarrayelems
%type<token> tokIdentVariable tokIdentModuleIdent tokVariableModuleVariable tokKeyword objectkey
%token<operator> tokAltOp tokUpdateOp tokDestAltOp tokOrOp tokAndOp tokCompareOp
%token<token> tokModule tokImport tokInclude tokDef tokAs tokLabel tokBreak
%token<token> tokNull tokTrue tokFalse
%token<token> tokIdent tokVariable tokModuleIdent tokModuleVariable
%token<token> tokIndex tokNumber tokFormat tokInvalid
%token<token> tokString tokStringStart tokStringQuery tokStringEnd
%token<token> tokIf tokThen tokElif tokElse tokEnd
%token<token> tokTry tokCatch tokReduce tokForeach
%token tokRecurse tokFuncDefPost tokTermPost tokEmptyCatch

%nonassoc tokFuncDefPost tokTermPost
%right '|'
%left ','
%right tokAltOp
%nonassoc tokUpdateOp
%left tokOrOp
%left tokAndOp
%nonassoc tokCompareOp
%left '+' '-'
%left '*' '/' '%'
%nonassoc tokAs tokIndex '.' '?' tokEmptyCatch
%nonassoc '[' tokTry tokCatch

%%

program
    : moduleheader programbody
    {
        if $1 != nil { $2.(*Query).Meta = $1.(*ConstObject) }
        yylex.(*lexer).result = $2.(*Query)
    }

moduleheader
    :
    {
        $$ = nil
    }
    | tokModule constobject ';'
    {
        $$ = $2;
    }

programbody
    : imports funcdefs
    {
        $$ = &Query{Imports: $1.([]*Import), FuncDefs: reverseFuncDef($2.([]*FuncDef)), Term: &Term{Type: TermTypeIdentity}}
    }
    | imports query
    {
        if $1 != nil { $2.(*Query).Imports = $1.([]*Import) }
        $$ = $2
    }

imports
    :
    {
        $$ = []*Import(nil)
    }
    | imports import
    {
        $$ = append($1.([]*Import), $2.(*Import))
    }

import
    : tokImport tokString tokAs tokIdentVariable metaopt ';'
    {
        $$ = &Import{ImportPath: $2, ImportAlias: $4, Meta: $5.(*ConstObject)}
    }
    | tokInclude tokString metaopt ';'
    {
        $$ = &Import{IncludePath: $2, Meta: $3.(*ConstObject)}
    }

metaopt
    :
    {
        $$ = (*ConstObject)(nil)
    }
    | constobject {}

funcdefs
    :
    {
        $$ = []*FuncDef(nil)
    }
    | funcdef funcdefs
    {
        $$ = append($2.([]*FuncDef), $1.(*FuncDef))
    }

funcdef
    : tokDef tokIdent ':' query ';'
    {
        $$ = &FuncDef{Name: $2, Body: $4.(*Query)}
    }
    | tokDef tokIdent '(' funcdefargs ')' ':' query ';'
    {
        $$ = &FuncDef{$2, $4.([]string), $7.(*Query)}
    }

funcdefargs
    : tokIdentVariable
    {
        $$ = []string{$1}
    }
    | funcdefargs ';' tokIdentVariable
    {
        $$ = append($1.([]string), $3)
    }

tokIdentVariable
    : tokIdent {}
    | tokVariable {}

query
    : funcdef query %prec tokFuncDefPost
    {
        $2.(*Query).FuncDefs = prependFuncDef($2.(*Query).FuncDefs, $1.(*FuncDef))
        $$ = $2
    }
    | query '|' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpPipe, Right: $3.(*Query)}
    }
    | term tokAs bindpatterns '|' query
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, &Suffix{Bind: &Bind{$3.([]*Pattern), $5.(*Query)}})
        $$ = &Query{Term: $1.(*Term)}
    }
    | tokReduce term tokAs pattern '(' query ';' query ')'
    {
        $$ = &Query{Term: &Term{Type: TermTypeReduce, Reduce: &Reduce{$2.(*Term), $4.(*Pattern), $6.(*Query), $8.(*Query)}}}
    }
    | tokForeach term tokAs pattern '(' query ';' query ')'
    {
        $$ = &Query{Term: &Term{Type: TermTypeForeach, Foreach: &Foreach{$2.(*Term), $4.(*Pattern), $6.(*Query), $8.(*Query), nil}}}
    }
    | tokForeach term tokAs pattern '(' query ';' query ';' query ')'
    {
        $$ = &Query{Term: &Term{Type: TermTypeForeach, Foreach: &Foreach{$2.(*Term), $4.(*Pattern), $6.(*Query), $8.(*Query), $10.(*Query)}}}
    }
    | tokIf query tokThen query ifelifs ifelse tokEnd
    {
        $$ = &Query{Term: &Term{Type: TermTypeIf, If: &If{$2.(*Query), $4.(*Query), $5.([]*IfElif), $6.(*Query)}}}
    }
    | tokTry query trycatch
    {
        $$ = &Query{Term: &Term{Type: TermTypeTry, Try: &Try{$2.(*Query), $3.(*Query)}}}
    }
    | tokLabel tokVariable '|' query
    {
        $$ = &Query{Term: &Term{Type: TermTypeLabel, Label: &Label{$2, $4.(*Query)}}}
    }
    | query '?'
    {
        if t := $1.(*Query).Term; t != nil {
            t.SuffixList = append(t.SuffixList, &Suffix{Optional: true})
        } else {
            $$ = &Query{Term: &Term{Type: TermTypeQuery, Query: $1.(*Query), SuffixList: []*Suffix{{Optional: true}}}}
        }
    }
    | query ',' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpComma, Right: $3.(*Query)}
    }
    | query tokAltOp query
    {
        $$ = &Query{Left: $1.(*Query), Op: $2, Right: $3.(*Query)}
    }
    | query tokUpdateOp query
    {
        $$ = &Query{Left: $1.(*Query), Op: $2, Right: $3.(*Query)}
    }
    | query tokOrOp query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpOr, Right: $3.(*Query)}
    }
    | query tokAndOp query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpAnd, Right: $3.(*Query)}
    }
    | query tokCompareOp query
    {
        $$ = &Query{Left: $1.(*Query), Op: $2, Right: $3.(*Query)}
    }
    | query '+' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpAdd, Right: $3.(*Query)}
    }
    | query '-' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpSub, Right: $3.(*Query)}
    }
    | query '*' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpMul, Right: $3.(*Query)}
    }
    | query '/' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpDiv, Right: $3.(*Query)}
    }
    | query '%' query
    {
        $$ = &Query{Left: $1.(*Query), Op: OpMod, Right: $3.(*Query)}
    }
    | term %prec tokTermPost
    {
        $$ = &Query{Term: $1.(*Term)}
    }

bindpatterns
    : pattern
    {
        $$ = []*Pattern{$1.(*Pattern)}
    }
    | bindpatterns tokDestAltOp pattern
    {
        $$ = append($1.([]*Pattern), $3.(*Pattern))
    }

pattern
    : tokVariable
    {
        $$ = &Pattern{Name: $1}
    }
    | '[' arraypatterns ']'
    {
        $$ = &Pattern{Array: $2.([]*Pattern)}
    }
    | '{' objectpatterns '}'
    {
        $$ = &Pattern{Object: $2.([]*PatternObject)}
    }

arraypatterns
    : pattern
    {
        $$ = []*Pattern{$1.(*Pattern)}
    }
    | arraypatterns ',' pattern
    {
        $$ = append($1.([]*Pattern), $3.(*Pattern))
    }

objectpatterns
    : objectpattern
    {
        $$ = []*PatternObject{$1.(*PatternObject)}
    }
    | objectpatterns ',' objectpattern
    {
        $$ = append($1.([]*PatternObject), $3.(*PatternObject))
    }

objectpattern
    : objectkey ':' pattern
    {
        $$ = &PatternObject{Key: $1, Val: $3.(*Pattern)}
    }
    | string ':' pattern
    {
        $$ = &PatternObject{KeyString: $1.(*String), Val: $3.(*Pattern)}
    }
    | '(' query ')' ':' pattern
    {
        $$ = &PatternObject{KeyQuery: $2.(*Query), Val: $5.(*Pattern)}
    }
    | tokVariable
    {
        $$ = &PatternObject{KeyOnly: $1}
    }

term
    : '.'
    {
        $$ = &Term{Type: TermTypeIdentity}
    }
    | tokRecurse
    {
        $$ = &Term{Type: TermTypeRecurse}
    }
    | tokIndex
    {
        $$ = &Term{Type: TermTypeIndex, Index: &Index{Name: $1}}
    }
    | '.' suffix
    {
        if $2.(*Suffix).Iter {
            $$ = &Term{Type: TermTypeIdentity, SuffixList: []*Suffix{$2.(*Suffix)}}
        } else {
            $$ = &Term{Type: TermTypeIndex, Index: $2.(*Suffix).Index}
        }
    }
    | '.' string
    {
        $$ = &Term{Type: TermTypeIndex, Index: &Index{Str: $2.(*String)}}
    }
    | tokNull
    {
        $$ = &Term{Type: TermTypeNull}
    }
    | tokTrue
    {
        $$ = &Term{Type: TermTypeTrue}
    }
    | tokFalse
    {
        $$ = &Term{Type: TermTypeFalse}
    }
    | tokIdentModuleIdent
    {
        $$ = &Term{Type: TermTypeFunc, Func: &Func{Name: $1}}
    }
    | tokIdentModuleIdent '(' args ')'
    {
        $$ = &Term{Type: TermTypeFunc, Func: &Func{Name: $1, Args: $3.([]*Query)}}
    }
    | tokVariableModuleVariable
    {
        $$ = &Term{Type: TermTypeFunc, Func: &Func{Name: $1}}
    }
    | tokNumber
    {
        $$ = &Term{Type: TermTypeNumber, Number: $1}
    }
    | tokFormat
    {
        $$ = &Term{Type: TermTypeFormat, Format: $1}
    }
    | tokFormat string
    {
        $$ = &Term{Type: TermTypeFormat, Format: $1, Str: $2.(*String)}
    }
    | string
    {
        $$ = &Term{Type: TermTypeString, Str: $1.(*String)}
    }
    | '(' query ')'
    {
        $$ = &Term{Type: TermTypeQuery, Query: $2.(*Query)}
    }
    | '+' term
    {
        $$ = &Term{Type: TermTypeUnary, Unary: &Unary{OpAdd, $2.(*Term)}}
    }
    | '-' term
    {
        $$ = &Term{Type: TermTypeUnary, Unary: &Unary{OpSub, $2.(*Term)}}
    }
    | '{' '}'
    {
        $$ = &Term{Type: TermTypeObject, Object: &Object{}}
    }
    | '{' objectkeyvals '}'
    {
        $$ = &Term{Type: TermTypeObject, Object: &Object{$2.([]*ObjectKeyVal)}}
    }
    | '{' objectkeyvals ',' '}'
    {
        $$ = &Term{Type: TermTypeObject, Object: &Object{$2.([]*ObjectKeyVal)}}
    }
    | '[' ']'
    {
        $$ = &Term{Type: TermTypeArray, Array: &Array{}}
    }
    | '[' query ']'
    {
        $$ = &Term{Type: TermTypeArray, Array: &Array{$2.(*Query)}}
    }
    | tokBreak tokVariable
    {
        $$ = &Term{Type: TermTypeBreak, Break: $2}
    }
    | term tokIndex
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, &Suffix{Index: &Index{Name: $2}})
    }
    | term suffix
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, $2.(*Suffix))
    }
    | term '?'
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, &Suffix{Optional: true})
    }
    | term '.' suffix
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, $3.(*Suffix))
    }
    | term '.' string
    {
        $1.(*Term).SuffixList = append($1.(*Term).SuffixList, &Suffix{Index: &Index{Str: $3.(*String)}})
    }

string
    : tokString
    {
        $$ = &String{Str: $1}
    }
    | tokStringStart stringparts tokStringEnd
    {
        $$ = &String{Queries: $2.([]*Query)}
    }

stringparts
    :
    {
        $$ = []*Query{}
    }
    | stringparts tokString
    {
        $$ = append($1.([]*Query), &Query{Term: &Term{Type: TermTypeString, Str: &String{Str: $2}}})
    }
    | stringparts tokStringQuery query ')'
    {
        yylex.(*lexer).inString = true
        $$ = append($1.([]*Query), &Query{Term: &Term{Type: TermTypeQuery, Query: $3.(*Query)}})
    }

tokIdentModuleIdent
    : tokIdent {}
    | tokModuleIdent {}

tokVariableModuleVariable
    : tokVariable {}
    | tokModuleVariable {}

suffix
    : '[' ']'
    {
        $$ = &Suffix{Iter: true}
    }
    | '[' query ']'
    {
        $$ = &Suffix{Index: &Index{Start: $2.(*Query)}}
    }
    | '[' query ':' ']'
    {
        $$ = &Suffix{Index: &Index{Start: $2.(*Query), IsSlice: true}}
    }
    | '[' ':' query ']'
    {
        $$ = &Suffix{Index: &Index{End: $3.(*Query), IsSlice: true}}
    }
    | '[' query ':' query ']'
    {
        $$ = &Suffix{Index: &Index{Start: $2.(*Query), End: $4.(*Query), IsSlice: true}}
    }

args
    : query
    {
        $$ = []*Query{$1.(*Query)}
    }
    | args ';' query
    {
        $$ = append($1.([]*Query), $3.(*Query))
    }

ifelifs
    :
    {
        $$ = []*IfElif(nil)
    }
    | ifelifs tokElif query tokThen query
    {
        $$ = append($1.([]*IfElif), &IfElif{$3.(*Query), $5.(*Query)})
    }

ifelse
    :
    {
        $$ = (*Query)(nil)
    }
    | tokElse query
    {
        $$ = $2
    }

trycatch
    : %prec tokEmptyCatch
    {
        $$ = (*Query)(nil)
    }
    | tokCatch query
    {
        $$ = $2
    }

objectkeyvals
    : objectkeyval
    {
        $$ = []*ObjectKeyVal{$1.(*ObjectKeyVal)}
    }
    | objectkeyvals ',' objectkeyval
    {
        $$ = append($1.([]*ObjectKeyVal), $3.(*ObjectKeyVal))
    }

objectkeyval
    : objectkey ':' objectval
    {
        $$ = &ObjectKeyVal{Key: $1, Val: $3.(*ObjectVal)}
    }
    | string ':' objectval
    {
        $$ = &ObjectKeyVal{KeyString: $1.(*String), Val: $3.(*ObjectVal)}
    }
    | '(' query ')' ':' objectval
    {
        $$ = &ObjectKeyVal{KeyQuery: $2.(*Query), Val: $5.(*ObjectVal)}
    }
    | objectkey
    {
        $$ = &ObjectKeyVal{KeyOnly: $1}
    }
    | string
    {
        $$ = &ObjectKeyVal{KeyOnlyString: $1.(*String)}
    }

objectkey
    : tokIdent {}
    | tokVariable {}
    | tokKeyword {}

objectval
    : term
    {
        $$ = &ObjectVal{[]*Query{{Term: $1.(*Term)}}}
    }
    | objectval '|' term
    {
        $$ = &ObjectVal{append($1.(*ObjectVal).Queries, &Query{Term: $3.(*Term)})}
    }

constterm
    : constobject
    {
        $$ = &ConstTerm{Object: $1.(*ConstObject)}
    }
    | constarray
    {
        $$ = &ConstTerm{Array: $1.(*ConstArray)}
    }
    | tokNumber
    {
        $$ = &ConstTerm{Number: $1}
    }
    | tokString
    {
        $$ = &ConstTerm{Str: $1}
    }
    | tokNull
    {
        $$ = &ConstTerm{Null: true}
    }
    | tokTrue
    {
        $$ = &ConstTerm{True: true}
    }
    | tokFalse
    {
        $$ = &ConstTerm{False: true}
    }

constobject
    : '{' '}'
    {
        $$ = &ConstObject{}
    }
    | '{' constobjectkeyvals '}'
    {
        $$ = &ConstObject{$2.([]*ConstObjectKeyVal)}
    }
    | '{' constobjectkeyvals ',' '}'
    {
        $$ = &ConstObject{$2.([]*ConstObjectKeyVal)}
    }

constobjectkeyvals
    : constobjectkeyval
    {
        $$ = []*ConstObjectKeyVal{$1.(*ConstObjectKeyVal)}
    }
    | constobjectkeyvals ',' constobjectkeyval
    {
        $$ = append($1.([]*ConstObjectKeyVal), $3.(*ConstObjectKeyVal))
    }

constobjectkeyval
    : tokIdent ':' constterm
    {
        $$ = &ConstObjectKeyVal{Key: $1, Val: $3.(*ConstTerm)}
    }
    | tokKeyword ':' constterm
    {
        $$ = &ConstObjectKeyVal{Key: $1, Val: $3.(*ConstTerm)}
    }
    | tokString ':' constterm
    {
        $$ = &ConstObjectKeyVal{KeyString: $1, Val: $3.(*ConstTerm)}
    }

constarray
    : '[' ']'
    {
        $$ = &ConstArray{}
    }
    | '[' constarrayelems ']'
    {
        $$ = &ConstArray{$2.([]*ConstTerm)}
    }

constarrayelems
    : constterm
    {
        $$ = []*ConstTerm{$1.(*ConstTerm)}
    }
    | constarrayelems ',' constterm
    {
        $$ = append($1.([]*ConstTerm), $3.(*ConstTerm))
    }

tokKeyword
    : tokOrOp {}
    | tokAndOp {}
    | tokModule {}
    | tokImport {}
    | tokInclude {}
    | tokDef {}
    | tokAs {}
    | tokLabel {}
    | tokBreak {}
    | tokNull {}
    | tokTrue {}
    | tokFalse {}
    | tokIf {}
    | tokThen {}
    | tokElif {}
    | tokElse {}
    | tokEnd {}
    | tokTry {}
    | tokCatch {}
    | tokReduce {}
    | tokForeach {}

%%
