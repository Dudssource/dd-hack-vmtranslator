package vm

import (
	"fmt"
	"strings"
)

const (
	// EOL character
	EOL    = '\n'
	INDENT = "    "
)

const (
	TRUE  uint16 = 0xFFFF
	FALSE uint16 = 0x0
)

// offsets for memory segments
const (
	SP   = "SP"
	LCL  = "LCL"
	ARG  = "ARG"
	THIS = "THIS"
	THAT = "THAT"
)

const (
	// base address for temp memory segment
	TEMP_REG = "R5"
	// used to help with pop
	POP_REG = "R13"
)

type codeWriter struct {
	w         *strings.Builder
	errorList []error
	eqIdx     int
	gtIdx     int
	ltIdx     int
}

func (o *codeWriter) add() {
	o.write("// add x + y")
	// Y=D
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// X=D
	o.pop()
	// X+Y
	o.write("@R14")
	o.write("D=D+M")
	// PUSH
	o.push()
}

func (o *codeWriter) sub() {
	o.write("// sub x - y")
	// Y=D
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// X=D
	o.pop()
	// X-Y
	o.write("@R14")
	o.write("D=M-D")
	// PUSH
	o.push()
}

func (o *codeWriter) neg() {
	o.write("// neg -y")
	// Y=D
	o.pop()
	// Y=-Y
	o.write("D=-D")
	// PUSH
	o.push()
}

func (o *codeWriter) eq() {
	o.write("// eq x==y")
	o.eqIdx++
	// R14=Y
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// D=X
	o.pop()
	// X-Y
	o.write("@R14")
	o.write("D=M-D")
	o.write(fmt.Sprintf("@EQ_%d", o.eqIdx))
	o.write("D;JEQ")
	o.write("D=0")
	o.push()
	o.write(fmt.Sprintf("(EQ_%d)", o.eqIdx))
	o.write("D=-1")
	o.push()
}

func (o *codeWriter) gt() {
	o.write("// gt x > y")
	o.gtIdx++
	// R14=Y
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// D=X
	o.pop()
	// X-Y
	o.write("@R14")
	o.write("D=M-D")
	o.write(fmt.Sprintf("@GT_%d", o.gtIdx))
	o.write("D;JGT")
	o.write("D=0")
	o.push()
	o.write(fmt.Sprintf("(GT_%d)", o.gtIdx))
	o.write("D=-1")
	o.push()
}

func (o *codeWriter) lt() {
	o.write("// lt x < y")
	o.ltIdx++
	// R14=Y
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// D=X
	o.pop()
	// X-Y
	o.write("@R14")
	o.write("D=M-D")
	o.write(fmt.Sprintf("@LT_%d", o.ltIdx))
	o.write("D;JLT")
	o.write("D=0")
	o.push()
	o.write(fmt.Sprintf("(LT_%d)", o.ltIdx))
	o.write("D=-1")
	o.push()
}

func (o *codeWriter) and() {
	o.write("// and x & y")
	// Y=D
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// X=D
	o.pop()
	// X & Y
	o.write("@R14")
	o.write("D=D&M")
	// PUSH
	o.push()
}

func (o *codeWriter) or() {
	o.write("// or x | y")
	// Y=D
	o.pop()
	o.write("@R14")
	o.write("M=D")
	// X=D
	o.pop()
	// X | Y
	o.write("@R14")
	o.write("D=D|M")
	// PUSH
	o.push()
}

func (o *codeWriter) not() {
	o.write("// or x | y")
	// Y=D (discarded)
	o.pop()
	// X=D
	o.pop()
	// !X
	o.write("D=!D")
	// PUSH
	o.push()
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
func (o *codeWriter) pushStatic(index, fileName string) {
	o.write("// push static " + index)
	o.write("@" + fileName + "." + index)
	o.write("D=M")
	o.push()
}

// popStatic index is turned into a variable with format fileName.i
func (o *codeWriter) popStatic(index, fileName string) {
	o.write("// pop static " + index)
	o.pop()
	o.write("@" + fileName + "." + index)
	o.write("M=D")
}

// pushTemp uses a fixed 8-entry segment, starting from address R5 plus index
func (o *codeWriter) pushTemp(index string) {
	o.write("// push temp " + index)
	o.pushDynamic(TEMP_REG, index)
}

// popTemp uses a fixed 8-entry segment, starting from address R5 (TEMP_REG) plus index
func (o *codeWriter) popTemp(index string) {
	o.write("// pop temp " + index)
	o.popDynamic(TEMP_REG, index)
}

// popPointer base address of this and that segments
func (o *codeWriter) pushPointer(index string) {
	o.write("// push pointer " + index)
	if index == "0" {
		o.pushThis("0")
	} else if index == "1" {
		o.pushThat("0")
	} else {
		o.errorList = append(o.errorList, fmt.Errorf("invalid push pointer index %s : should be 0 or 1", index))
	}
}

// popPointer base address of this and that segments
func (o *codeWriter) popPointer(index string) {
	o.write("// pop pointer " + index)
	if index == "0" {
		o.popThis("0")
	} else if index == "1" {
		o.popThat("0")
	} else {
		o.errorList = append(o.errorList, fmt.Errorf("invalid pop pointer index %s : should be 0 or 1", index))
	}
}

// pushDynamic used by local,argument,this, that and temp memory segments
func (o *codeWriter) pushDynamic(segmentPointer, index string) {
	o.write("@" + segmentPointer)
	o.write("D=M")
	o.write("@" + index)
	o.write("A=D+A")
	o.write("D=M")
	o.push()
}

// push generic stack push logic
// D register should contain the input
func (o *codeWriter) push() {
	o.write("@SP")
	o.write("A=M")
	o.write("M=D")
	o.write("@SP")
	o.write("M=M+1")
}

// popDynamic used by local,argument,this and that memory segments
func (o *codeWriter) popDynamic(segmentPointer, index string) {
	o.write("@" + segmentPointer)
	o.write("D=M")
	o.write("@" + index)
	o.write("D=D+A")
	o.write("@" + POP_REG)
	o.write("M=D")
	o.pop()
	o.write("@" + POP_REG)
	o.write("A=M")
	o.write("M=D")
}

// pop generic stack pop logic
// outputs into D register
func (o *codeWriter) pop() {
	o.write("@SP")
	o.write("M=M-1")
	o.write("A=M")
	o.write("D=M")
}

// write encapsulate inner builder write logic, including EOL
func (o *codeWriter) write(instruction string) {
	o.w.WriteString(instruction)
	o.w.WriteRune(EOL)
}
