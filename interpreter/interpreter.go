//Package interpreter execute hub commands
package interpreter

import (
	"errors"
	"fmt"

	"github.com/PMoneda/hub/ast"
	"github.com/PMoneda/hub/lang"
	"github.com/PMoneda/hub/lexer"
	"github.com/PMoneda/hub/utils"
)

//Interpreter instance to run hub code
type Interpreter struct {
	lexer          *lexer.Lexer
	executionState int
	root           ast.Tree
}

//Run execute a hub script
func (interpreter *Interpreter) Run(lexer *lexer.Lexer) {
	interpreter.lexer = lexer
	interpreter.root = *new(ast.Tree)
	interpreter.root.Value = "BEGIN"
	for lexer.HasNext() {
		var execRoot ast.Tree

		token := lexer.Next()
		if token == "\\EOF\\" {
			break
		}
		interpreter.stmt(&execRoot, token)
		interpreter.root.AppendChild(&execRoot)
	}

}

//Print ast
func (interpreter *Interpreter) Print() {
	interpreter.print(&interpreter.root)
}

//Print ast
func (interpreter *Interpreter) print(root *ast.Tree) {
	root.Print(func(value interface{}) {
		switch v := value.(type) {
		case lang.Object:
			fmt.Println(v.ToString())
			break
		default:
			fmt.Println(v)
		}

	})
}

//exec is the initial point of execution and starts with de first token on hub script
func (interpreter *Interpreter) stmt(parent *ast.Tree, token string) error {
	switch token {
	case "var":
		interpreter.varStmt(parent)
		return nil
	case "print":
		interpreter.printStmt(parent)
		return nil
	case "read":
		interpreter.readStmt(parent)
		return nil
	case "if":
		interpreter.readIf(parent)
		return nil
	default:
		return errors.New("end")

	}
}
func (interpreter *Interpreter) printExecInfo() {
	tok := interpreter.lexer.GetTokenInfo()
	fmt.Printf("error at line %d token %d near %s\n", tok.Line, tok.Index, tok.Value)
}
func (interpreter *Interpreter) matchKeyword(next string, keyword string) (string, error) {
	if next != keyword {
		interpreter.printExecInfo()
		err := errors.New("Expected: " + keyword + " got: " + next)
		//return "", err
		panic(err.Error())
	}
	return next, nil //OK
}

func (interpreter *Interpreter) matchIdent() (string, error) {
	next := interpreter.lexer.Next()
	if !interpreter.lexer.IsIdent(next) {
		interpreter.printExecInfo()
		err := errors.New("Expected Identifier  got: " + next)
		return "", err
	}
	return next, nil //OK
}
func (interpreter *Interpreter) matchEOF(token string) error {
	if token == "\\EOF\\" {
		tok := interpreter.lexer.Previous()
		err := errors.New("Unexpected end of file line: " + string(tok.Line))
		return err
	}
	return nil
}

func (interpreter *Interpreter) matchEndOfCommand(token string) {
	if token == ";" {
		return
	}
	interpreter.printExecInfo()
	panic("; expected")
}

func (interpreter *Interpreter) matchEOL(token string) {
	if token == "\n" {
		return
	}
	interpreter.printExecInfo()
	panic("End of line expected")
}

func (interpreter *Interpreter) readStmt(root *ast.Tree) {
	var readStmt ast.ReadStmt
	readStmt.Op = "read"
	root.Value = readStmt
	var idenNode ast.Tree
	interpreter.identStmt(&idenNode)
	next := interpreter.lexer.Next()
	interpreter.matchEndOfCommand(next)
	root.AppendChild(&idenNode)
}
func (interpreter *Interpreter) stmtBlock(root *ast.Tree, token string) {
	interpreter.matchKeyword("{", token)
	for token != "}" {
		var left ast.Tree
		token = interpreter.lexer.Next()

		stmt := interpreter.stmt(&left, token)
		if stmt == nil {
			root.AppendChild(&left)
		}
	}
	token = interpreter.lexer.Current()
	interpreter.matchKeyword("}", token)
}
func (interpreter *Interpreter) readIf(root *ast.Tree) {
	root.Value = "if"
	//var right ast.Tree
	var cond ast.Tree
	interpreter.exprStmt(&cond)
	token := interpreter.lexer.Current()
	interpreter.stmtBlock(&cond, token)
	//token = interpreter.lexer.Next()
	if token == "else" {
		token = interpreter.lexer.Next()
		interpreter.stmtBlock(&cond, token)
		root.AppendChild(&cond)
	} else {
		root.AppendChild(&cond)
		//interpreter.stmt(root, token)
	}

}

func (interpreter *Interpreter) printStmt(root *ast.Tree) {
	var printStmt ast.PrintStmt
	printStmt.Op = "print"
	root.Value = printStmt
	var expNode ast.Tree
	interpreter.exprStmt(&expNode)
	token := interpreter.lexer.Current()
	interpreter.matchEndOfCommand(token)
	root.AppendChild(&expNode)
}

//Statement rule for list of commands
func (interpreter *Interpreter) varStmt(root *ast.Tree) {
	var decl ast.DeclVar
	var iden ast.Tree
	var exp ast.Tree
	interpreter.identStmt(&iden)
	next := interpreter.lexer.Next()
	_, err1 := interpreter.matchKeyword(next, "=")
	if err1 != nil {
		panic(err1.Error())
	}
	decl.Op = "="
	interpreter.exprStmt(&exp)
	next = interpreter.lexer.Current()
	interpreter.matchEndOfCommand(next)
	root.Value = decl
	root.AppendChild(&iden)
	root.AppendChild(&exp)
}
func (interpreter *Interpreter) identStmt(parent *ast.Tree) {
	id, err := interpreter.matchIdent()
	if err != nil {
		panic(err.Error())
	}
	var ident = lang.BuildPointer(id)
	parent.Value = ident

}

func (interpreter *Interpreter) exprStmt(root *ast.Tree) {

	interpreter.E(root)
}

//BuildObject based on token
func (interpreter *Interpreter) BuildObject(token string) lang.Object {
	lex := interpreter.lexer
	if lex.IsNumber(token) {
		return lang.BuildNumber(token)
	} else if lex.IsString(token) {
		return lang.BuildString(token)
	} else if lex.IsIdent(token) {
		return lang.BuildPointer(token)
	} else if lex.IsOperator(token) {
		return lang.BuildOperator(token)
	} else if lex.IsBoolean(token) {
		return lang.BuildBoolean(token)
	}
	return nil
}

//E reflects a expression rule
func (interpreter *Interpreter) E(root *ast.Tree) lang.Object {
	exp := interpreter.convExpToStack()
	root.Value = exp
	//evaluate the expression
	return nil
}

//Converts a expression to stack of operations
func (interpreter *Interpreter) convExpToStack() utils.Stack {
	lex := interpreter.lexer
	token := lex.Next()
	var terms utils.Stack
	var ops utils.Stack
	for !lex.IsBlockDelimiter(token) && !lex.IsCommandDelimiter(token) && token != "\\EOF\\" {
		if !lex.IsOperator(token) && !lex.IsParenhesis(token) {
			terms.Push(interpreter.BuildObject(token))
		} else if token == "(" {
			ops.Push(token)
		} else if token == ")" {
			top := ops.Pop()
			for top != "(" {
				terms.Push(top)
				if ops.IsEmpty() {
					interpreter.printExecInfo()
					panic("Invalid Expression: missing '('")
				} else {
					top = ops.Pop()
				}
			}
		} else {
			currOp := interpreter.BuildObject(token).(lang.Op)
			if !ops.IsEmpty() {
				switch v := ops.Top().(type) {
				case lang.Object:
					op := v.(lang.Op)
					if op.HighPriority(currOp) {
						terms.Push(ops.Pop())
					}
					break
				}
			}
			ops.Push(currOp)
		}
		token = lex.Next()
	}
	for !ops.IsEmpty() {
		switch v := ops.Pop().(type) {
		case lang.Object:
			terms.Push(v)
			break
		default:
			interpreter.printExecInfo()
			panic(fmt.Sprintf("Invalid expression  at: %v", v))
		}

	}
	return terms
}
