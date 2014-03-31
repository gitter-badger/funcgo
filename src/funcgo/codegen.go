//////
// This file is part of the Funcgo compiler.
//
// Copyright (c) 2012,2013 Eamonn O'Brien-Strain All rights
// reserved. This program and the accompanying materials are made
// available under the terms of the Eclipse Public License v1.0 which
// accompanies this distribution, and is available at
// http://www.eclipse.org/legal/epl-v10.html
//
// Contributors:
// Eamonn O'Brien-Strain e@obrain.com - initial author
//////

package  funcgo/codegen
import (
        s "clojure/string"
)

func listStr(item...) {
	str("(", s.join(" ", item), ")")
}

func vecStr(item...) {
	str("[", s.join(" ", item), "]")
}

func infix(expression) {
	expression
} (left, operator, right) {
	listStr(operator, left, right)
}

codeGenerator :=  {
	SOURCEFILE:     func(header, body) {str(header, body, "\n")},
	PACKAGECLAUSE:  func(imported, importDecl) {
		str("(ns ", imported, " (:gen-class)", importDecl, ")(set! *warn-on-reflection* true)\n\n")
	},
	IMPORTDECL:     func(importSpecs...) {apply(str, importSpecs)},
	IMPORTSPEC:     func(identifier, imported) {
		str("\n  (:require [", imported, " :as ", identifier, "])")
	} (imported) {
		str("\n  (:require [", imported, " :as ", last(imported s.split /\./), "])")
	},
	PRECEDENCE0: infix,
	PRECEDENCE1: infix,
	PRECEDENCE2: infix,
	PRECEDENCE3: infix,
	PRECEDENCE4: infix,
	PRECEDENCE5: infix,
	IFELSEEXPR: func(condition, exprs) {
		listStr("when", condition, exprs)
	} (condition, block1, block2) {
		listStr("if", condition, block1, block2)
	},
	FORRANGE: func(identifier, seq, expressions) {
		str("(doseq [", identifier, " ", seq, "] ", expressions, ")")
	},
	FORLAZY: func(identifier, seq, expressions) {
		str("(for [", identifier, " ", seq, "] ", expressions, ")")
	} (identifier, seq, condition, expressions) {
		str("(for [", identifier, " ", seq, "] :when ", condition, " ", expressions, ")")
	},
	FORTIMES: func(identifier, count, expressions) {
		str("(dotimes [", identifier, " ", count, "] ", expressions, ")")
	},
	TRYEXPR: func(expressions, catches) {
		listStr("try", expressions, catches)
	} (expressions, catches, finally) {
		listStr("try", expressions, catches, finally)
	},
	CATCHES: func(catches...){
		s.join(" ", catches)
	},
	CATCH: func(typ, exception, expressions) {
		listStr("catch", typ, exception, expressions)
	},
	FINALLY: func(expressions) {
		listStr("finally", expressions)
	},
	NEW: func(symbol){str(symbol, ".")},
	SHORTVARDECL:   func(identifier, expression) {
		listStr("def", identifier, expression)
	},
	VARDECL:   func(identifier, expression) {
		listStr("def", identifier, expression)
	} (identifier, typ, expression) {
		listStr("def", "^" str typ, identifier, expression)
	},
	FUNCTIONCALL:    func(function) {
		listStr(function)
	} (function, call) {
		listStr(function, call)
	},
	EXPRESSIONLIST: func(expr0, exprRest...){
		s.join(" ", expr0 cons exprRest)
	},
	EXPRESSIONS: func(expr0, exprRest...){
		s.join(" ", expr0 cons exprRest)
	},
	EXPRESSIONSXXX: func(expr0, exprRest...){
		s.join(" ", expr0 cons exprRest)
	},
	CONSTS:  func(consts...) {
		"\n" s.join consts
	},
	BLOCK: func (expr){
		expr
	} (expr0, exprRest...) {
		str(
			"(do ",
			(" " s.join (expr0 cons exprRest)),
			")"
		)
	},
	INDEXED: func(xs, i){ listStr("nth", xs, i) },
	WITHCONST: func(xs...){
		const(
			consts = butlast(xs)
			expressions = last(xs)
		)
		str("(let [",
			" " s.join consts,
			"] ",
			expressions,
			")")
	},
	CONST: func(identifier, expression) {
		str(identifier, " ", expression)
	},
	VECDESTRUCT: vecStr,
	DICTDESTRUCT: func(elems...) {
		str('{', (" " s.join elems), "}")
	}, 
        DICTDESTRUCTELEM: func(destruct, label) {
		str(destruct, " ", label)
	},
	VARIADICDESTRUCT:  func(destruct) {str("& ", destruct)},
	SYMBOL: func(identifier){
		identifier
	} (pkg, identifier) {
		str(pkg, "/", identifier)
	},
	BINARYOP: identity,
	MULOP: identity,
	ADDOP: identity,
	RELOP: identity,
	FUNCTIONDECL:   func(identifier, function) {
		listStr("defn", identifier, function)
	},
	FUNCTIONLIT:    func(function) {listStr("fn", function)},
	FUNCTIONPARTS:  func(functionpart...) {
		str("(",
			") (" s.join functionpart,
			")")
	},
	FUNCTIONPART0:  func(expression) {
		"[] " str expression
	} (typ, expression) {
		str("^", typ, " [] ", expression)
	},
	VFUNCTIONPART0:  func(variadic, expression) {
		str("[", variadic, "] ", expression)
	} (variadic, typ, expression) {
		str("^", typ, " [", variadic, "] ", expression)
	},
	FUNCTIONPARTN:  func(parameters, expression) {
		str("[", parameters, "] ", expression)
	} (parameters, typ, expression) {
		str("^", typ, " [", parameters, "] ", expression)
	},
	VFUNCTIONPARTN: func(parameters, variadic, expression) {
		str("[", parameters, " ", variadic, "] ", expression)
	} (parameters, variadic, typ, expression) {
		str("^", typ, " [", parameters, " ", variadic, "] ", expression)
	},
	PARAMETERS:     func(arg0, argsRest...) {
		" " s.join (arg0 cons argsRest)
	},
	VARIADIC:       func(parameter) {str("& ", parameter)},
	VECLIT:         func() {
		"[]"
	} (expressions...) {
		str(
			"[",
			" " s.join expressions,
			"]"
		)
	},
	DICTLIT:        func(dictElems...) {apply(str, dictElems)},
	DICTELEMENT:    func(key, value) {str(key, " ", value, " ")},
	LABEL:          func(s) {
		str(":", s.replace(s.lowerCase(s), /_/, "-"))
	},
	IDENTIFIER:     func(s) {
		s.replace(s, /\p{Ll}\p{Lu}/, func(s){
			str(first(s), "-", s.lowerCase(last(s)))
		})
	},
	TYPEDIDENTIFIER: func(identifier, typ) {
		str(`^`, typ, " ", identifier)
	},
	IMPORTED:         func(idf0, idfRest...){
		"." s.join (idf0 cons idfRest)
	},
	DECIMALLIT:    identity,
	BIGINTLIT:     str,
	BIGFLOATLIT:   str,
	FLOATLIT:      str,
	DECIMALS:      identity,
	EXPONENT:      str,
	REGEX:         func(s){str(`#"`, s, `"`)},
	INTERPRETEDSTRINGLIT: func(s){str(`"`, s, `"`)},
	CLOJUREESCAPE: identity,
	LITTLEUVALUE:  func(d1,d2,d3,d4){str(`\u`,d1,d2,d3,d4)},
	OCTALBYTEVALUE:  func(d1,d2,d3){str(`\o`,d1,d2,d3)},
	UNICODECHAR:   func(s){`\` str s},
	NEWLINECHAR:   func(){`\newline`},
	SPACECHAR:     func(){`\space`},
	BACKSPACECHAR: func(){`\backspace`},
	RETURNCHAR:    func(){`\return`},
	TABCHAR:       func(){`\tab`},
	BACKSLASHCHAR: func(){`\\`},
	SQUOTECHAR:    func(){`\'`},
	DQUOTECHAR:    func(){`\"`},
	HEXDIGIT:      identity,
	OCTALDIGIT:    identity,
	RAWSTRINGLIT:  func(s){str(`"`, s.escape(s, charEscapeString), `"`)},
	DASHIDENTIFIER: func(s){ "-" str s },
	ISIDENTIFIER:   func(initial, identifier) {
		str( s.lowerCase(initial), identifier, "?")
	},
	EQUALS: func() { "=" },
	AND: func() { "and" },
	OR: func() { "or" },
	MUTIDENTIFIER:  func(initial, identifier) {
		str( s.lowerCase(initial), identifier, "!")
	},
	ESCAPEDIDENTIFIER:  func(identifier) { identifier },
	UNARYEXPR: func(operator, expression) { listStr(operator, expression) },
	NOTEQ:       func(){ "not=" },
	BITXOR:      func(){ "bit-xor" },
	BITNOT:      func(){ "bit-not" },
	SHIFTLEFT:   func(){ "bit-shift-left" },
	SHIFTRIGHT:  func(){ "bit-shift-right" },
	NOT:         func(){ "not" },
	DEREF: func(expression) { str("@", expression) },
	SYNTAXQUOTE: func(expression)     { str("`", expression) },
	UNQUOTE: func(expression)         { str("~", expression) },
	UNQUOTESPLICING: func(expression) { str("~@", expression) },
	JAVAFIELD:      func(expression, identifier) {
		listStr(".", expression, identifier)
	},
	JAVASTATIC:      func(parts...) {"/" s.join parts},
	TYPE:            func(parts...) {"." s.join parts},
	JAVAMETHODCALL: func(expression, identifier) {
		str("(. ", expression, " (", identifier, "))")
	} (expression, identifier, call) {
		str("(. ", expression, " (", identifier, " ", call, "))")
	}
}