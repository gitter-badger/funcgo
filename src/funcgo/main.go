//////
// This file is part of the Funcgo compiler.
//
// Copyright (c) 2014 Eamonn O'Brien-Strain All rights
// reserved. This program and the accompanying materials are made
// available under the terms of the Eclipse Public License v1.0 which
// accompanies this distribution, and is available at
// http://www.eclipse.org/legal/epl-v10.html
//
// Contributors:
// Eamonn O'Brien-Strain e@obrain.com - initial author
//////

// This file contains the entry point for the standalone version of
// the Funcgo compile and is called by the Leiningen plugin.

package  main
import (
        "clojure/java/io"
        "clojure/pprint"
        "clojure/string"
        "clojure/tools/cli"
        "funcgo/core"
)
import type (
	java.io.{BufferedWriter, File, StringWriter, IOException}
	jline.console.ConsoleReader
)

commandLineOptions := [
        ["-r", "--repl",  "start a Funcgo interactive console"],
        ["-s", "--sync", "No asynchronous channel constructs"],
        ["-n", "--nodes", "print out the parse tree that the parser produces"],
        ["-u", "--ugly",  "do not pretty-print the Clojure"],
        ["-f", "--force", "Force compiling even if not out-of-date"],
        ["-a", "--ambiguity",  "print out all matched parse trees to diagnose ambiguity"],
        ["-h", "--help",  "print help"]
]

// A version of pprint that preserves type hints.
// See https://groups.google.com/forum/#!topic/clojure/5LRmPXutah8
func prettyPrint(obj, writer) {
	origDispatch := \pprint/*print-pprint-dispatch*\          // */ for emacs
	pprint.withPprintDispatch(
		func(o) {
			if met := meta(o); met {
				print("^")
				if count(met) == 1 {
					if met(TAG) {
						origDispatch(met(TAG))
					} else {
						if met(PRIVATE) == true {
							origDispatch(PRIVATE)
						} else {
							origDispatch(met)
						}
					}
				} else {
					origDispatch(met)
				}
				print(" ")
				pprint.pprintNewline(FILL)
			}
			origDispatch(o)
		},
		pprint.pprint(obj, writer)
	)
}

func writePrettyTo(cljText, writer BufferedWriter) {
	for expr := range readString( str("[", cljText, "]")) {
		prettyPrint(expr, writer)
		writer->newLine()
	}
	writer->close()
}


func compileExpression(inPath, fgoText) {
	cljText   := core.Parse(inPath, fgoText, EXPR)
	strWriter := new StringWriter()
	writer    := new BufferedWriter(strWriter)
	cljText  writePrettyTo  writer
	strWriter->toString()
}

func newConsoleReader() {
	consoleReader := new ConsoleReader()
	consoleReader->setPrompt("fgo=>     ")
	consoleReader
}

func repl(){
	consoleReader ConsoleReader := newConsoleReader()
	loop(){
		fgoText := consoleReader->readLine()
		if !string.isBlank(fgoText) {
			try{
				cljText := first(core.Parse("repl.go", fgoText, EXPR))
				println("Clojure: ", cljText)
				println("Result:  ", eval(readString(cljText)))
			} catch Exception e {
				println(e)
			}
			println()
		}
		if fgoText != nil {
			recur()
		}
	}
}

func CompileString(inPath, fgoText) {
	cljText   := core.Parse(inPath, fgoText)
	strWriter := new StringWriter()
	writer    := new BufferedWriter(strWriter)
	cljText  writePrettyTo  writer
	strWriter->toString()
}

func compileFile(inFile File, root File, opts) {
	splitRoot := reMatches(/([^\.]+)(\.[a-z]+)?(\.gos?)/, inFile->getPath)
	if !isNil(splitRoot) {
		[_, inPath, suffixExtra, suffix] := splitRoot
		compileFile(
			inFile,
			root,
			inPath  str  suffix,
			opts,
			if isNil(suffixExtra) {""} else {suffixExtra}
		)
	}
} (inFile File, root File, inPath, opts, suffixExtra) {
	outFile := io.file(string.replace(inPath, /\.go(s?)$/, ".clj$1"  str  suffixExtra))
	if opts(FORCE) || outFile->lastModified() < inFile->lastModified() {
		prefixLen := root->getAbsolutePath()->length()
		relative  := subs(inFile->getAbsolutePath(), prefixLen + 1)
		println("  ", relative, "...")
		{
			fgoText        := slurp(inFile)
			lines          := count(func{ $1 == '\n' }  filter  fgoText)
			start          := if suffixExtra == "" { SOURCEFILE } else { NONPKGFILE }
			try {
				beginTime := System::currentTimeMillis()
				cljText String := core.Parse(
					relative,
					fgoText,
					start,
					opts(NODES), opts(SYNC), opts(AMBIGUITY)
				)
				duration := System::currentTimeMillis() - beginTime
				// TODO(eob) open using with-open
				writer         := io.writer(outFile)

				writer->write(str(";; Compiled from ", inFile, "\n"))
				if opts(UGLY) {
					writer->write(cljText)
					writer->close()
				} else {
					cljText  writePrettyTo  writer
				}
				if outFile->length() == 0 {
					outFile->delete()
					println("\t\tERROR: No output created.")
				} else {
					println("\t\t-->",
						outFile->getPath(),
						int(1000.0*lines/duration),
						"lines/s")
					if (outFile->length) / (inFile->length) < 0.4 {
						println("WARNING: Output file is only",
							int(100 * (outFile->length)
								/ (inFile->length)),
							"% the size of the input file")
					}
				}
			} catch IOException e {
				println("Parsing ", relative, " failed:\n", e->getMessage())
			}
		}
  }
}

func compileTree(root File, opts) {
	println(root->getName())
	for f := range fileSeq(root) {
		inFile File := f
		try {
			compileFile(inFile, root, opts)
		} catch IOException e {
			println(e->getMessage())
		} catch Exception e {
			e->printStackTrace()
		}
	}
}

func printError(cmdLine) {
	println()
	if cmdLine(ERRORS) {
		println(cmdLine(ERRORS))
	}
	println("USAGE:  fgoc [options] path ...")
	println("options:")
	println(cmdLine(SUMMARY))
}

// Convert Funcgo files to clojure files, using the commandLineOptions
// to parse the arguments.
func Compile(args...) {
	cmdLine   := args  cli.parseOpts  commandLineOptions
	otherArgs := cmdLine(ARGUMENTS)
	opts      := cmdLine(OPTIONS)
	here      := io.file(".")

	if cmdLine(ERRORS) || opts(HELP){
		println(cmdLine(SUMMARY))
	}else{
		if not(seq(otherArgs)) {
			println("Missing directory or file argument.")
			printError(cmdLine)
		} else {
			// file arguments
			for arg := range otherArgs {
				if file := io.file(arg); file->isDirectory {
					compileTree(file, opts)
				} else {
					try {
						compileFile(file, here, opts)
					} catch Exception e {
						println("\n", e->getMessage())
					}
				}
			}
		}
		if opts(REPL) {
			repl()
		}
	}
}

// Entry point for stand-alone compiler. Usage is the same as for the
// Compile function.
func _main(args...) {
	Compile(...args)
}
