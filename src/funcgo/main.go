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

package  funcgo.main
import (
	io clojure.java.io
        pprint clojure.pprint
        string clojure.string
	core funcgo.core
)

func compileFile(inFile) {
	const(
		inPath = inFile->getPath()
		outFile = io.file(string.replace(inPath, /\.go$/, ".clj"))
	)
	if outFile->lastModified() < inFile->lastModified() {
		const(
			clj = core.funcgoParse(slurp(inFile))
			// TODO(eob) open using with-open
			writer = io.writer(outFile)
		)
		println(inPath)
		writer->write(str(";; Compiled from ", inFile, "\n"))
		for expr := range readString( str("[", clj, "]")) {
			pprint.pprint(expr, writer)
			writer->newLine()
		}
		writer->close()
		println("  -->", outFile->getPath())
		if (outFile->length) / (inFile->length) < 0.5 {
			println("WARNING: Output file is only",
				int(100 * (outFile->length) / (inFile->length)),
				"% the size of the input file")
		}
	}
}

 // Convert funcgo to clojure
func _main(&args) {
	if not(seq(args)) {
		for f := range fileSeq(io.file(".")) {
			try {
				if f->getName()->endsWith(".go") { 
					compileFile(f)
				}
			} catch Exception e {
				println("\n", e->getMessage())
			}
		}
	}else{
		for arg := range args {
			compileFile(io.file(arg))
		}
	}
}
