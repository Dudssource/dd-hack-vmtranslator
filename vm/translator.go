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
package vm

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// clean tabs and comments
	cleanUpRegex = regexp.MustCompile(`[\t]+|//.*$`)
)

type Translator struct {
	cw *codeWriter
}

func Translate(srcPath string) error {

	// clean up src path
	srcPath = strings.TrimRight(srcPath, string(os.PathSeparator))

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat %s : %w", srcPath, err)
	}

	var (
		errorList      = make([]error, 0)
		outputfileName string
		cw             = &codeWriter{w: &strings.Builder{}, functionReturnMap: make(map[string]int)}
		tr             = &Translator{cw: cw}
	)

	if srcInfo.IsDir() {

		// directory name
		outputfileName = filepath.Base(srcPath)

		// add bootstrap instructions, only required for multi-file input
		cw.bootstrap()

		// all matching vm files within directory
		matches, err := filepath.Glob(fmt.Sprintf("%s/*.vm", strings.TrimRight(srcPath, string(os.PathSeparator))))
		if err != nil {
			return err
		}

		// add path
		srcPath += "/"

		// translate all files
		for _, vmFile := range matches {
			translateVMFile(vmFile, tr, &errorList)
		}

	} else {

		// file name (without extension)
		outputfileName = strings.Split(filepath.Base(srcPath), ".")[0]

		// translate single file
		translateVMFile(srcPath, tr, &errorList)
	}

	if len(errorList) > 0 {
		return errors.Join(errorList...)
	}

	var (
		// target path
		dstPath = filepath.Join(filepath.Dir(srcPath), outputfileName+".asm")

		// convert result to binary
		binaryOutput = []byte(tr.cw.w.String())
	)

	// sanity check
	if len(binaryOutput) == 0 {
		return errors.New("HACK VM Translator found no instructions found to be processed")
	}

	// write output (need to remove extra EOL at the end of output)
	if err := os.WriteFile(dstPath, binaryOutput[:len(binaryOutput)-1], 0766); err != nil {
		return fmt.Errorf("hackvmtranslator : write output : %s : %w", dstPath, err)
	}

	// ok
	log.Printf("HACK VM Translator finished successfully, output to %s\n", dstPath)

	// no errors
	return nil
}

func translateVMFile(srcPath string, tr *Translator, errorList *[]error) {

	// open src file
	file, err := os.Open(srcPath)
	if err != nil {
		*errorList = append(*errorList, err)
		return
	}

	var (
		fileName = strings.ReplaceAll(filepath.Base(srcPath), ".vm", "")
		scanner  = bufio.NewScanner(file)
	)

	// set current file
	tr.cw.fileName = fileName

	for lineNo := 1; scanner.Scan(); lineNo++ {

		// read line, cleaning up white spaces, tabs and comments
		line := strings.TrimSpace(cleanUpRegex.ReplaceAllString(scanner.Text(), ""))

		// extract symbols
		parts := strings.Split(line, " ")

		switch parts[0] {
		case "eq":
			tr.cw.eq()
		case "add":
			tr.cw.add()
		case "sub":
			tr.cw.sub()
		case "neg":
			tr.cw.neg()
		case "not":
			tr.cw.not()
		case "and":
			tr.cw.and()
		case "or":
			tr.cw.or()
		case "gt":
			tr.cw.gt()
		case "lt":
			tr.cw.lt()
		case "push":

			if len(parts) < 3 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid push %s", lineNo, line))
				continue
			}

			segment := parts[1]
			index := parts[2]

			switch segment {
			case "local":
				tr.cw.pushLocal(index)
			case "argument":
				tr.cw.pushArgument(index)
			case "this":
				tr.cw.pushThis(index)
			case "that":
				tr.cw.pushThat(index)
			case "temp":
				tr.cw.pushTemp(index)
			case "constant":
				tr.cw.pushConstant(index)
			case "pointer":
				tr.cw.pushPointer(index)
			case "static":
				tr.cw.pushStatic(index)
			}

		case "pop":

			if len(parts) < 3 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid pop %s", lineNo, line))
				continue
			}

			segment := parts[1]
			index := parts[2]

			switch segment {
			case "local":
				tr.cw.popLocal(index)
			case "argument":
				tr.cw.popArgument(index)
			case "this":
				tr.cw.popThis(index)
			case "that":
				tr.cw.popThat(index)
			case "temp":
				tr.cw.popTemp(index)
			case "pointer":
				tr.cw.popPointer(index)
			case "static":
				tr.cw.popStatic(index)
			}

		case "label":
			if len(parts) < 2 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid label %s", lineNo, line))
				continue
			}
			tr.cw.label(parts[1])

		case "goto":
			if len(parts) < 2 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid goto %s", lineNo, line))
				continue
			}
			tr.cw.goTo(parts[1])

		case "if-goto":
			if len(parts) < 2 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid if-goto %s", lineNo, line))
				continue
			}
			tr.cw.ifGoTo(parts[1])

		case "function":
			if len(parts) < 3 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid function %s", lineNo, line))
				continue
			}
			tr.cw.function(parts[1], parts[2])

		case "call":
			if len(parts) < 3 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid call %s", lineNo, line))
				continue
			}
			tr.cw.call(parts[1], parts[2])

		case "return":
			if len(parts) != 1 {
				*errorList = append(*errorList, fmt.Errorf("%d : invalid function %s", lineNo, line))
				continue
			}
			tr.cw.returns()
		}
	}
}
