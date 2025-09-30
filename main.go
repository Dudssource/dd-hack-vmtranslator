// This file is part of DD HACK VM Translator.
// Copyright (C) 2025-2025 Eduardo <dudssource@gmail.com>
//
// HACK VM Translator is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// HACK VM Translator is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with HACK VM Translator.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"log"
	"os"

	"github.com/Dudssource/dd-hack-vmtranslator/vm"
)

func main() {

	// parse args
	args := os.Args

	// validate src
	if len(args) != 2 {
		log.Fatalln(`Usage of hackvmtranslator:
  VMTranslator myProg/FileName.vm
  VMTranslator myProg/
`)
	}

	// run translator
	if err := vm.Translate(args[1]); err != nil {
		log.Fatalf("Found errors:\n%s\n", err.Error())
	}
}
