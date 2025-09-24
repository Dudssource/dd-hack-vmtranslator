package vm

import (
	"fmt"
	"strconv"
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
	o.write("// not x")
	o.write("@SP")
	o.write("AM=M-1") // SP--
	o.write("A=A-1")  // X = RAM[SP-1]
	o.write("M=!M")   // RAM[SP-1] = !X
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
	o.write(fmt.Sprintf("@%d", o.toInt(index)+5))
	o.write("D=M")
	o.push()
}

// popTemp uses a fixed 8-entry segment, starting from address R5 (TEMP_REG) plus index
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
