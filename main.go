// This file is part of DD HACK VM Translator.
// Copyright (C) 2025-2025 Eduardo <dudssource@gmail.com>
//
// DD HACK Assembler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// DD HACK Assembler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with DD HACK Assembler.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"flag"
	"log"

	"github.com/Dudssource/dd-hack-vmtranslator/vm"
)

func main() {

	// parse args
	flag.Args()
	src := flag.String("s", "", "HACK VM translator source file")
	out := flag.String("o", "", "HACK VM translator output file")
	dbg := flag.Bool("x", false, "Debug translation")
	flag.Parse()

	// validate src/out
	if len(*src) == 0 || len(*out) == 0 {
		flag.Usage()
		log.Fatalln("Missing required arguments")
	}

	// run translator
	if err := vm.Translate(src, out, dbg); err != nil {
		log.Fatalf("Found errors:\n%s\n", err.Error())
	}
}
