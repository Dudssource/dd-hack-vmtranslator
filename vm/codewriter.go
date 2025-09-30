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
	"fmt"
	"strconv"
	"strings"
)

const (
	// EOL character
	EOL = '\n'
)

// offsets for memory segments
const (
	LCL  = "LCL"
	ARG  = "ARG"
	THIS = "THIS"
	THAT = "THAT"
)

const (
	SYS_INIT_LABEL = "Sys.init"
)

type codeWriter struct {
	w                 *strings.Builder
	eqIdx             int
	gtIdx             int
	ltIdx             int
	fileName          string // current file name (without extension)
	functionReturnMap map[string]int
}

func (o *codeWriter) bootstrap() {
	o.write("@256")
	o.write("D=A")
	o.write("@SP")
	o.write("M=D")
	// call
	o.call(SYS_INIT_LABEL, "0")
}

func (o *codeWriter) label(name string) {
	o.write("(" + name + ")")
}

func (o *codeWriter) goTo(name string) {
	o.write("@" + name)
	o.write("0;JMP") // goto
}

func (o *codeWriter) ifGoTo(name string) {
	o.pop()
	o.write("@" + name)
	o.write("D;JNE") // if D != 0 goto
}

func (o *codeWriter) call(name, nArgs string) {

	// return address
	o.functionReturnMap[name]++
	label := fmt.Sprintf("%s$ret.%d", name, o.functionReturnMap[name])
	o.write("@" + label)
	o.write("D=A")
	o.push()

	// memory segments
	o.write("@" + LCL)
	o.write("D=M")
	o.push()
	o.write("@" + ARG)
	o.write("D=M")
	o.push()
	o.write("@" + THIS)
	o.write("D=M")
	o.push()
	o.write("@" + THAT)
	o.write("D=M")
	o.push()

	// reposition args
	o.write("@SP")
	o.write("D=M-1")
	for idx := o.toInt(nArgs) + 4; idx > 0; idx-- {
		o.write("D=D-1")
	}
	o.write("@" + ARG)
	o.write("M=D")

	// reposition LCL
	o.write("@SP")
	o.write("D=M")
	o.write("@" + LCL)
	o.write("M=D")

	// goto Foo.mult
	o.goTo(name)

	// write goto ret
	o.write("(" + label + ")")
}

func (o *codeWriter) function(name, nArgs string) {
	// (Foo.bar)
	o.write("(" + name + ")")

	// number of args
	val := o.toInt(nArgs) - 1

	// initialize local variables
	for idx := val; idx >= 0; idx-- {
		o.write("D=0")
		o.push()
	}
}

func (o *codeWriter) returns() {
	// end frame (LCL)
	o.write("@" + LCL)
	o.write("D=M")
	o.write("@R13") // end frame
	o.write("M=D")
	o.write("@5")
	o.write("A=D-A")
	o.write("D=M")
	o.write("@R14") // ret addr
	o.write("M=D")

	o.pop()
	o.write("@" + ARG)
	o.write("A=M")
	o.write("M=D")

	o.write("@" + ARG)
	o.write("D=M+1")
	o.write("@SP")
	o.write("M=D")

	o.write("@R13")
	o.write("AM=M-1")
	o.write("D=M")
	o.write("@" + THAT)
	o.write("M=D")

	o.write("@R13")
	o.write("AM=M-1")
	o.write("D=M")
	o.write("@" + THIS)
	o.write("M=D")

	o.write("@R13")
	o.write("AM=M-1")
	o.write("D=M")
	o.write("@" + ARG)
	o.write("M=D")

	o.write("@R13")
	o.write("AM=M-1")
	o.write("D=M")
	o.write("@" + LCL)
	o.write("M=D")

	o.write("@R14")
	o.write("A=M")
	o.write("0;JMP")
}

func (o *codeWriter) add() {
	o.write("// add x + y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("M=D+M")  // RAM[SP-1] = X + Y
}

func (o *codeWriter) sub() {
	o.write("// sub x - y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("M=M-D")  // RAM[SP-1] = X - Y
}

func (o *codeWriter) neg() {
	o.write("// neg -y")
	o.write("@SP")
	o.write("A=M-1") // Y = RAM[SP-1]
	o.write("M=-M")  // Y = -Y
}

func (o *codeWriter) eq() {
	o.write("// eq x==y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("D=M-D")  // D = X-Y
	o.write("M=0")    // RAM[SP-1] = 0
	o.eqIdx++
	o.write(fmt.Sprintf("@NEQ_%d", o.eqIdx))
	o.write("D;JNE") // if D != 0 goto NEQ_N
	o.write("@SP")
	o.write("A=M-1")
	o.write("M=-1")
	o.write(fmt.Sprintf("(NEQ_%d)", o.eqIdx))
}

func (o *codeWriter) gt() {
	o.write("// gt x > y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("D=M-D")  // D = X-Y
	o.write("M=0")    // RAM[SP-1] = 0
	o.gtIdx++
	o.write(fmt.Sprintf("@NGT_%d", o.gtIdx))
	o.write("D;JLE") // if D <= 0 goto NGT_N
	o.write("@SP")
	o.write("A=M-1")
	o.write("M=-1")
	o.write(fmt.Sprintf("(NGT_%d)", o.gtIdx))
}

func (o *codeWriter) lt() {
	o.write("// lt x < y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("D=M-D")  // D = X - Y
	o.write("M=0")    // RAM[SP-1] = 0
	o.ltIdx++
	o.write(fmt.Sprintf("@NLT_%d", o.ltIdx))
	o.write("D;JGE") // if D >= 0 goto NLT_N
	o.write("@SP")
	o.write("A=M-1")
	o.write("M=-1")
	o.write(fmt.Sprintf("(NLT_%d)", o.ltIdx))
}

func (o *codeWriter) and() {
	o.write("// and x & y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("M=D&M")  // RAM[SP-1] = D & M
}

func (o *codeWriter) or() {
	o.write("// or x | y")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("D=M")    // Y = RAM[SP]
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("M=D|M")  // RAM[SP-1] = D | M
}

func (o *codeWriter) not() {
	o.write("// not y")
	o.write("@SP")
	o.write("A=M-1") // SP--
	o.write("M=!M")  // RAM[SP-1] = !M
}

func (*codeWriter) toInt(index string) int {
	n, _ := strconv.ParseInt(index, 10, 16)
	return int(n)
}

func (o *codeWriter) pushLocal(index string) {
	o.write("// push local " + index)
	o.pushDynamic(LCL, index)
}

func (o *codeWriter) popLocal(index string) {
	o.write("// pop local " + index)
	o.popDynamic(LCL, index)
}

func (o *codeWriter) pushArgument(index string) {
	o.write("// push argument " + index)
	o.pushDynamic(ARG, index)
}

func (o *codeWriter) popArgument(index string) {
	o.write("// pop argument " + index)
	o.popDynamic(ARG, index)
}

func (o *codeWriter) pushThis(index string) {
	o.write("// push this " + index)
	o.pushDynamic(THIS, index)
}

func (o *codeWriter) popThis(index string) {
	o.write("// pop this " + index)
	o.popDynamic(THIS, index)
}

func (o *codeWriter) pushThat(index string) {
	o.write("// push that " + index)
	o.pushDynamic(THAT, index)
}

func (o *codeWriter) popThat(index string) {
	o.write("// pop that " + index)
	o.popDynamic(THAT, index)
}

// pushConstant value is constant, no pop available for constant
func (o *codeWriter) pushConstant(value string) {
	o.write("// push constant " + value)
	o.write("@" + value)
	o.write("D=A")
	o.push()
}

// pushStatic index is turned into a variable with format fileName.i
func (o *codeWriter) pushStatic(index string) {
	o.write("// push static " + index)
	o.write("@" + o.fileName + "." + index)
	o.write("D=M")
	o.push()
}

// popStatic index is turned into a variable with format fileName.i
func (o *codeWriter) popStatic(index string) {
	o.write("// pop static " + index)
	o.pop()
	o.write("@" + o.fileName + "." + index)
	o.write("M=D")
}

// pushTemp uses a fixed 8-entry segment, starting from address R5 plus index
func (o *codeWriter) pushTemp(index string) {
	o.write("// push temp " + index)
	o.write(fmt.Sprintf("@%d", o.toInt(index)+5))
	o.write("D=M")
	o.push()
}

// popTemp uses a fixed 8-entry segment, starting from address R5 plus index
func (o *codeWriter) popTemp(index string) {
	o.write("// pop temp " + index)
	o.pop()
	o.write(fmt.Sprintf("@%d", o.toInt(index)+5))
	o.write("M=D")
}

// popPointer base address of this and that segments
func (o *codeWriter) pushPointer(index string) {
	o.write("// push pointer " + index)
	if index == "0" {
		o.write("@" + THIS)
	} else {
		o.write("@" + THAT)
	}
	o.write("D=M")
	o.push()
}

// popPointer base address of this and that segments
func (o *codeWriter) popPointer(index string) {
	o.write("// pop pointer " + index)
	o.pop()
	if index == "0" {
		o.write("@" + THIS)
	} else {
		o.write("@" + THAT)
	}
	o.write("M=D")
}

// pushDynamic used by local,argument,this, that and temp memory segments
func (o *codeWriter) pushDynamic(segmentPointer, index string) {
	o.write("@" + segmentPointer)
	o.write("A=M")
	for range o.toInt(index) {
		o.write("A=A+1")
	}
	o.write("D=M")
	o.push()
}

// push generic stack push logic
// D register should contain the input
func (o *codeWriter) push() {
	o.write("@SP")
	o.write("AM=M+1")
	o.write("A=A-1")
	o.write("M=D")
}

// popDynamic used by local,argument,this and that memory segments
func (o *codeWriter) popDynamic(segmentPointer, index string) {
	o.pop()
	o.write("@" + segmentPointer)
	o.write("A=M")
	for range o.toInt(index) {
		o.write("A=A+1")
	}
	o.write("M=D")
}

// pop generic stack pop logic
// outputs into D register
func (o *codeWriter) pop() {
	o.write("@SP")
	o.write("AM=M-1")
	o.write("D=M")
}

// write encapsulate inner builder write logic, including EOL
func (o *codeWriter) write(instruction string) {
	o.w.WriteString(instruction)
	o.w.WriteRune(EOL)
}
