package opcodes

import (
	"fmt"

	"github.com/PMoneda/hub/lang"
)

//Sum represents opcode sum
//sum r0 r1 r0 Ex: execute r0 = r0 + r1
type Sum struct {
	Op1    interface{}
	Op2    interface{}
	Result interface{}
}

//ToString print Sum command
func (opcode Sum) ToString() string {
	op := "sum"
	switch v := opcode.Op1.(type) {
	case lang.Pointer:
		op += " $" + v.ToString()
		break
	case lang.Object:
		op += " #" + v.ToString()
		break
	default:
		op += fmt.Sprintf(" %v", v)
		break
	}
	switch v := opcode.Op2.(type) {
	case lang.Pointer:
		op += " $" + v.ToString()
		break
	case lang.Object:
		op += " #" + v.ToString()
		break
	case string:
		op += fmt.Sprintf(" %v", v)
		break
	default:
		op += fmt.Sprintf(" #%v", v)
		break
	}

	switch v := opcode.Result.(type) {
	case lang.Pointer:
		op += " $" + v.ToString()
		break
	case lang.Object:
		op += " #" + v.ToString()
		break
	case string:
		op += fmt.Sprintf(" %v", v)
		break
	default:
		op += fmt.Sprintf(" #%v", v)
		break
	}

	return op
}

//Execute execute Sum command
func (opcode Sum) Execute() {
	fmt.Println("Execute Sum")
}
