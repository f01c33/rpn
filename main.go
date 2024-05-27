package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"os"
	"strings"
)

type Type int

const (
	Number Type = iota
	Variable
	String
	Assignment
	Code
)

var (
	inFile   string
	outFile  string
	debug    bool
	Zero     = big.NewFloat(0)
	One      = big.NewFloat(1)
	Mode     = "dec"
	Vertical = false
	Exit     = false
	// x of interally eXecutable, c of constant, m of macro, a of assignment
	keyWords = map[string]string{
		// Arithmetic Operators

		"+":   "x", // Add
		"-":   "x", // Subtract
		"*":   "x", // Multiply
		"/":   "x", // Divide
		"cla": "x", // Clear the stack and variables
		"clr": "x", // Clear the stack
		"clv": "x", // Clear the variables
		"!":   "x", // Boolean NOT
		"%":   "x", // Modulus
		"++":  "x", // Increment
		"--":  "x", // Decrement

		// Bitwise Operators

		"&":  "x", // Bitwise AND
		"|":  "x", // Bitwise OR
		"^":  "x", // Bitwise XOR
		"~":  "x", // Bitwise NOT
		"<<": "x", // Bitwise shift left
		">>": "x", // Bitwise shift right

		// Boolean Operators

		"&&": "x", // Boolean AND
		"||": "x", // Boolean OR
		"^^": "x", // Boolean XOR

		// Comparison Operators

		"!=": "x", // Not equal to
		"<":  "x", // Less than
		"<=": "x", // Less than or equal to
		"==": "x", // Equal to
		">":  "x", // Greater than
		">=": "x", // Greater than or equal to

		// Trigonometric Functions

		"acos": "x", // Arc Cosine
		"asin": "x", // Arc Sine
		"atan": "x", // Arc Tangent
		"cos":  "x", // Cosine
		"cosh": "x", // Hyperbolic Cosine
		"sin":  "x", // Sine
		"sinh": "x", // Hyperbolic Sine
		"tanh": "x", // Hyperbolic tangent

		// Numeric Utilities

		"ceil":  "x", // Ceiling
		"floor": "x", // Floor
		"round": "x", // Round
		"ip":    "x", // Integer part
		"fp":    "x", // Floating part
		"sign":  "x", // Push -1, 0, or 0 depending on the sign
		"abs":   "x", // Absolute value
		"max":   "x", // Max
		"min":   "x", // Min

		// Display Modes

		"hex": "x", // Switch display mode to hexadecimal
		"dec": "x", // Switch display mode to decimal (default)
		"bin": "x", // Switch display mode to binary
		"oct": "x", // Switch display mode to octal

		// Constants

		"e":    "c", // Push e
		"pi":   "c", // Push Pi
		"rand": "c", // Generate a random number

		// Mathematic Functions

		"exp":  "x", // Exponentiation
		"fact": "x", // Factorial
		"sqrt": "x", // Square Root
		"ln":   "x", // Natural Logarithm
		"log":  "x", // Logarithm
		"pow":  "x", // Raise a number to a power

		// Networking

		"hnl": "x", // Host to network long
		"hns": "x", // Host to network short
		"nhl": "x", // Network to host long
		"nhs": "x", // Network to host short

		// Stack Manipulation

		"pick":   "x", // Pick the -n'th item from the stack
		"repeat": "x", // Repeat an operation n times, e.g. '3 repeat +'
		"depth":  "x", // Push the current stack depth
		"drop":   "x", // Drops the top item from the stack
		"dropn":  "x", // Drops n items from the stack
		"dup":    "x", // Duplicates the top stack item
		"dupn":   "x", // Duplicates the top n stack items in order
		"roll":   "x", // Roll the stack upwards by n
		"rolld":  "x", // Roll the stack downwards by n
		"stack":  "x", // Toggles stack display from horizontal to vertical
		"swap":   "x", // Swap the top 2 stack items

		// Macros and Variables

		"macro": "m", // Defines a macro, e.g. 'macro kib 1024 *'
		"=":     "a", // Assigns a variable, e.g. '1024 x='

		// Other

		"help":  "x", // Print the help message
		"exit":  "x", // Exit the calculator
		"debug": "x", // toggle debug mode
	}
)

// rpc -in stdin
// for interactive mode
func init() {
	// flag.StringVar(&inFile, "in", "stdin", "Select the input file (stdin, for example)")
	p := "stdin"
	flag.StringVar(&inFile, "in", p, "Select the input file (stdin, for example)")
	flag.StringVar(&outFile, "out", "stdout", "Select the output file (stdout, for example)")
	flag.BoolVar(&debug, "g", false, "Debug mode")
	flag.Parse()
}

func getFiles() (in *os.File, out *os.File) {
	var err error
	if inFile == "stdin" {
		in = os.Stdin
	} else {
		in, err = os.Open(inFile)
		if err != nil {
			panic(err)
		}
	}
	if outFile == "stdout" {
		out = os.Stdout
	} else {
		out, err = os.Open(outFile)
		if err != nil {
			panic(err)
		}
	}
	return
}

func getBase(s string) int {
	if len(s) == 0 {
		// what
		return -1
	}
	if s[0] == 0 && len(s) > 3 {
		switch s[1] {
		case 'x':
			return 16
		case 'd':
			return 10
		case 'o':
			return 8
		case 'b':
			return 2
		}
	}
	tmpF := float32(42)
	tmpI := int(42)
	_, err := fmt.Sscanf(s, "%f", &tmpF)
	_, err2 := fmt.Sscanf(s, "%i", &tmpI)
	if err != nil && err2 != nil {
		return 0
	}
	return 10
}

type Var struct {
	Type Type       // type of the thing
	V    string     // Variable
	F    *big.Float // Float
	B    []byte     // "String"
	Code []Var      // Code
}

func (v Var) String() string {
	switch v.Type {
	case Number:
		return v.F.String() + ":Number"
	case Variable:
		return v.V + ":Variable"
	case Code:
		if len(v.Code) != 0 {
			return v.V + ":Code" + fmt.Sprint(v.Code)
		}
		return v.V + ":Code"
	case Assignment:
		return v.V + ":Assignment"
	}
	return ""
}

func isKeyword(s string) bool {
	_, ok := keyWords[s]
	return ok
}

func Parse(lex []string) (stack []Var, vars map[string]Var) {
	stack = make([]Var, 0)
	vars = make(map[string]Var, 0)
	for i := range lex {
		if debug {
			fmt.Fprintf(os.Stderr, "Parsing %s\n", lex[i])
		}
		base := getBase(lex[i])
		// lex[i] is string if base == 0
		if base == 0 {
			// if last char is =
			if lex[i][:len(lex[i])-1] == "=" {
				// add it to the variables
				vars[lex[i][:len(lex[i])-1]] = Var{Type: Code, V: "=", F: Zero}
				stack = append(stack, Var{Type: Assignment, V: lex[i][:len(lex[i])-1], F: Zero})
				if debug {
					fmt.Fprintln(os.Stderr, "variable assignment:", lex[i])
				}
				continue
			} else if isKeyword(lex[i]) {
				if debug {
					fmt.Fprintln(os.Stderr, "keyword:", lex[i])
				}
				stack = append(stack, Var{Type: Code, V: lex[i], F: Zero})
				continue
			}
			if debug {
				fmt.Fprintln(os.Stderr, "variable:", lex[i])
			}
			stack = append(stack, Var{Type: Variable, V: lex[i], F: Zero})
			continue
			// if wrong base log the occurence
		} else if debug && base == -1 {
			fmt.Fprintln(os.Stderr, fmt.Errorf("error parsing: \"%v\", at the lexeme number %v", lex[i], i))
		}
		if debug {
			fmt.Fprintln(os.Stderr, "Number:", lex[i])
		}
		f, b, err := big.NewFloat(0).Parse(lex[i], base)
		if debug {
			fmt.Fprintln(os.Stderr, "base:", b)
		}
		if err != nil {
			continue
			// panic(err)
		}
		stack = append(stack, Var{Type: Number, F: f})
	}
	return
}

func remove(stack []Var, pos, removeN int) []Var {
	return append(stack[:pos-removeN], stack[pos:]...)
}

func Eval(stack []Var, vars map[string]Var, ip int) ([]Var, map[string]Var, int) {
	// apply function to previous stack values
	for i := ip; i < len(stack); i++ {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in ", stack[i], r)
			}
		}()
		if debug {
			fmt.Fprintln(os.Stderr, "Evaluating: ", stack[i], "with variables: ", vars)
		}
		switch stack[i].Type {
		case Number:
			continue
		case Variable:
			var ok bool
			prev := vars[stack[i].V]
			stack[i], ok = vars[stack[i].V]
			if ok {
				i -= 1
			} else {
				stack[i] = prev
				fmt.Fprint(os.Stderr, "The variable doesn't exist\n")
			}
			continue
		case Code:
			if isKeyword(stack[i].V) {
				stack[i].F = big.NewFloat(0)
				switch stack[i].V {
				case "debug":
					fmt.Fprintf(os.Stderr, "Toggling debug mode\n")
					debug = !debug
					remove(stack, i+1, 1)
					i -= 1
				case "+":
					if debug {
						fmt.Fprintf(os.Stderr, "%v + %v\n", stack[i-2].F, stack[i-1].F)
					}
					stack[i].F.Add(stack[i-2].F, stack[i-1].F)
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "-":
					if debug {
						fmt.Fprintf(os.Stderr, "%v - %v\n", stack[i-2].F, stack[i-1].F)
					}
					stack[i].F.Sub(stack[i-2].F, stack[i-1].F)
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2

				case "*":
					if debug {
						fmt.Fprintf(os.Stderr, "%v * %v\n", stack[i-2].F, stack[i-1].F)
					}
					stack[i].F.Mul(stack[i-2].F, stack[i-1].F)
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "/":
					if debug {
						fmt.Fprintf(os.Stderr, "%v / %v\n", stack[i-2].F, stack[i-1].F)
					}
					stack[i].F.Quo(stack[i-2].F, stack[i-1].F)
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "cla": // Clear the stack and variables
					if debug {
						fmt.Fprintf(os.Stderr, "Clearing stack and variables\n")
					}
					stack = make([]Var, 0)
					vars = make(map[string]Var, 0)
					i = 0
				case "clr": // Clear the stack
					if debug {
						fmt.Fprintf(os.Stderr, "Clearing the stack\n")
					}
					stack = make([]Var, 0)
					i = 0
				case "clv": // Clear the variables
					if debug {
						fmt.Fprintf(os.Stderr, "Clearing the variables\n")
					}
					vars = make(map[string]Var, 0)
				case "!": // Boolean NOT
					if debug {
						fmt.Fprintf(os.Stderr, "!%v\n", stack[i-1])
					}
					stack[i-1].F.Neg(stack[i-1].F)
					stack = remove(stack, i+1, 1)
					i -= 1
				case "%": // Modulus
					if debug {
						fmt.Fprintf(os.Stderr, "%v mod %v\n", stack[i-2].F, stack[i-1].F)
					}
					div, _ := stack[i-2].F.Int(nil)
					rem, _ := stack[i-1].F.Int(nil)
					stack[i].F.SetInt(big.NewInt(0).Rem(div, rem))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "++": // Increment
					if debug {
						fmt.Fprintf(os.Stderr, "%v++\n", stack[i-1].F)
					}
					stack[i-1].F.Add(stack[i-1].F, big.NewFloat(1))
					stack = remove(stack, i+1, 1)
					i -= 1
				case "--": // Decrement

					if debug {
						fmt.Fprintf(os.Stderr, "%v--\n", stack[i-1].F)
					}
					stack[i-1].F.Sub(stack[i-1].F, big.NewFloat(1))
					stack = remove(stack, i+1, 1)
					i -= 1
				case "&": // Bitwise AND
					if debug {
						fmt.Fprintf(os.Stderr, "%v & %v\n", stack[i-2].F, stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					b, _ := stack[i-1].F.Int(nil)
					stack[i].F.SetInt(big.NewInt(0).And(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "|": // Bitwise OR

					if debug {
						fmt.Fprintf(os.Stderr, "%v | %v\n", stack[i-2].F, stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					b, _ := stack[i-1].F.Int(nil)
					stack[i].F.SetInt(big.NewInt(0).Or(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "^": // Bitwise XOR

					if debug {
						fmt.Fprintf(os.Stderr, "%v ^ %v\n", stack[i-2].F, stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					b, _ := stack[i-1].F.Int(nil)
					stack[i].F.SetInt(big.NewInt(0).Xor(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "~": // Bitwise NOT
					if debug {
						fmt.Fprintf(os.Stderr, "~%v\n", stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					stack[i].F.SetInt(big.NewInt(0).Not(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "<<": // Bitwise shift left
					if debug {
						fmt.Fprintf(os.Stderr, "%v << %v\n", stack[i-2].F, stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					b, _ := stack[i-1].F.Uint64()
					stack[i].F.SetInt(big.NewInt(0).Lsh(a, uint(b)))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case ">>": // Bitwise shift right
					if debug {
						fmt.Fprintf(os.Stderr, "%v >> %v\n", stack[i-2].F, stack[i-1].F)
					}
					a, _ := stack[i-2].F.Int(nil)
					b, _ := stack[i-1].F.Uint64()
					stack[i].F.SetInt(big.NewInt(0).Rsh(a, uint(b)))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "!=": // Not equal to
					if debug {
						fmt.Fprintf(os.Stderr, "%v != %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = One
					case 1:
						stack[i].F = One
					case 0:
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "==": // Equal to
					if debug {
						fmt.Fprintf(os.Stderr, "%v == %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = Zero
					case 1:
						stack[i].F = Zero
					case 0:
						stack[i].F = One
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "<": // Less than
					if debug {
						fmt.Fprintf(os.Stderr, "%v < %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = One
					case 1:
						stack[i].F = Zero
					case 0:
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "<=": // Less than or equal to
					if debug {
						fmt.Fprintf(os.Stderr, "%v <= %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = One
					case 1:
						stack[i].F = Zero
					case 0:
						stack[i].F = One
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case ">": // Greater than
					if debug {
						fmt.Fprintf(os.Stderr, "%v > %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = Zero
					case 1:
						stack[i].F = One
					case 0:
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case ">=": // Greater than or equal to
					if debug {
						fmt.Fprintf(os.Stderr, "%v >= %v\n", stack[i-1], stack[i-2])
					}
					switch stack[i-1].F.Cmp(stack[i-2].F) {
					case -1:
						stack[i].F = Zero
					case 1:
						stack[i].F = One
					case 0:
						stack[i].F = One
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "&&": // Boolean AND
					// ok for boolean comparison to lose precision
					if debug {
						fmt.Fprintf(os.Stderr, "%v && %v\n", stack[i-1], stack[i-2])
					}
					a, _ := stack[i-2].F.Uint64()
					b, _ := stack[i-1].F.Uint64()
					if a != 0 && b != 0 {
						stack[i].F = One
					} else {
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "||": // Boolean OR
					if debug {
						fmt.Fprintf(os.Stderr, "%v || %v\n", stack[i-1], stack[i-2])
					}
					a, _ := stack[i-2].F.Uint64()
					b, _ := stack[i-1].F.Uint64()
					if a != 0 || b != 0 {
						stack[i].F = One
					} else {
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "^^": // Boolean XOR
					if debug {
						fmt.Fprintf(os.Stderr, "%v ^^ %v\n", stack[i-1], stack[i-2])
					}
					// !a && b || a && !b
					a, _ := stack[i-2].F.Uint64()
					ab := a != 0
					b, _ := stack[i-1].F.Uint64()
					bb := b != 0
					if !ab && bb || ab && !bb {
						stack[i].F = One
					} else {
						stack[i].F = Zero
					}
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "acos": // Arc Cosine
					if debug {
						fmt.Fprintf(os.Stderr, "acos(%v)\n", stack[i-1])
					}
					// i'm sorry for i have sined, i didn't implement a bignum version of the trig functions, it would take too long
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Acos(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "asin": // Arc Sine
					if debug {
						fmt.Fprintf(os.Stderr, "asin(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Asin(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "atan": // Arc Tangent
					if debug {
						fmt.Fprintf(os.Stderr, "atan(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Atan(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "cos": // Cosine
					if debug {
						fmt.Fprintf(os.Stderr, "cos(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Cos(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "cosh": // Hyperbolic Cosine
					if debug {
						fmt.Fprintf(os.Stderr, "cosh(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Cosh(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "sin": // Sine
					if debug {
						fmt.Fprintf(os.Stderr, "sin(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Sin(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "sinh": // Hyperbolic Sine
					if debug {
						fmt.Fprintf(os.Stderr, "sinh(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Sinh(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "tanh": // Hyperbolic tangent
					if debug {
						fmt.Fprintf(os.Stderr, "tanh(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Tanh(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "hex": // Switch display mode to hexadecimal
					if debug {
						fmt.Fprintf(os.Stderr, "mode changed to hex\n")
					}
					Mode = "hex"
					stack = remove(stack, i+1, 1)
					i -= 1
				case "dec": // Switch display mode to decimal (default)
					if debug {
						fmt.Fprintf(os.Stderr, "mode changed to dec\n")
					}
					Mode = "dec"
					stack = remove(stack, i+1, 1)
					i -= 1
				case "bin": // Switch display mode to binary
					if debug {
						fmt.Fprintf(os.Stderr, "mode changed to bin\n")
					}
					Mode = "bin"
					stack = remove(stack, i+1, 1)
					i -= 1
				case "oct": // Switch display mode to octal
					if debug {
						fmt.Fprintf(os.Stderr, "mode changed to oct\n")
					}
					Mode = "oct"
					stack = remove(stack, i+1, 1)
					i -= 1
				case "e": // Push e
					if debug {
						fmt.Fprintf(os.Stderr, "pushed e\n")
					}
					stack[i].F = big.NewFloat(math.E)
					stack[i].Type = Number
				case "pi": // Push Pi
					if debug {
						fmt.Fprintf(os.Stderr, "pushed pi\n")
					}
					stack[i].F = big.NewFloat(math.Pi)
					stack[i].Type = Number
				case "rand": // Generate a random number [0.0,1.0)
					if debug {
						fmt.Fprintf(os.Stderr, "pushed random number\n")
					}
					stack[i].F = big.NewFloat(rand.Float64())
					stack[i].Type = Number
				case "pow": // Raise a number to a power
					if debug {
						fmt.Fprintf(os.Stderr, "%v**%v\n", stack[i-1], stack[i-2])
					}
					a, _ := stack[i-2].F.Float64()
					b, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Pow(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "**": // Raise a number to a power
					if debug {
						fmt.Fprintf(os.Stderr, "%v**%v\n", stack[i-1], stack[i-2])
					}
					a, _ := stack[i-2].F.Float64()
					b, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Pow(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "exp": // Exponentiation
					if debug {
						fmt.Fprintf(os.Stderr, "%v**%v\n", stack[i-1], stack[i-2])
					}
					a, _ := stack[i-2].F.Float64()
					b, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Pow(a, b))
					stack[i].Type = Number
					stack = remove(stack, i, 2)
					i -= 2
				case "fact": // Factorial
					if debug {
						fmt.Fprintf(os.Stderr, "%v!\n", stack[i-1])
					}
					// the lack of .Copy on big.Int made me angry, didn't memoize it, and found mulrange, happy little accident
					a, _ := stack[i-1].F.Int64()
					stack[i].F = big.NewFloat(0).SetInt(big.NewInt(1).MulRange(1, a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "sqrt": // Square Root
					if debug {
						fmt.Fprintf(os.Stderr, "sqrt(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Pow(a, 0.5))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "ln": // Natural Logarithm
					if debug {
						fmt.Fprintf(os.Stderr, "ln(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Log(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "log": // Logarithm
					if debug {
						fmt.Fprintf(os.Stderr, "log(%v)\n", stack[i-1])
					}
					a, _ := stack[i-1].F.Float64()
					stack[i].F = big.NewFloat(math.Log10(a))
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "hnl": // Host to network long
					if debug {
						fmt.Fprintf(os.Stderr, "hnl(%v)\n", stack[i-1])
					}
					buf := new(bytes.Buffer)
					a, _ := stack[i-1].F.Int64()
					err := binary.Write(buf, binary.BigEndian, a)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
					}
					stack[i].F = stack[i-1].F
					stack[i].B = buf.Bytes()
					stack[i].Type = String
					stack = remove(stack, i, 1)
					i -= 1
				case "hns": // Host to network short
					if debug {
						fmt.Fprintf(os.Stderr, "hns(%v)\n", stack[i-1])
					}
					buf := new(bytes.Buffer)
					a, _ := stack[i-1].F.Int64()
					err := binary.Write(buf, binary.BigEndian, int32(a))
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
					}
					stack[i].F = stack[i-1].F
					stack[i].B = buf.Bytes()
					stack[i].Type = String
					stack = remove(stack, i, 1)
					i -= 1
				case "nhl": // Network to host long
					if debug {
						fmt.Fprintf(os.Stderr, "nhl(%v)\n", stack[i-1])
					}
					stack[i] = stack[i-1]
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "nhs": // Network to host short
					if debug {
						fmt.Fprintf(os.Stderr, "nhs(%v)\n", stack[i-1])
					}
					stack[i] = stack[i-1]
					stack[i].Type = Number
					stack = remove(stack, i, 1)
					i -= 1
				case "pick": // Pick the -n'th item from the stack
					if debug {
						fmt.Fprintf(os.Stderr, "pick(%v)\n", stack[i-1])
					}
					j, _ := stack[i-1].F.Int64()
					stack[i] = stack[j]
					stack = remove(stack, i, 1)
					i -= 1
				case "repeat": // Repeat an operation n times, e.g. '3 repeat +'
					if debug {
						fmt.Fprintf(os.Stderr, "for i = 0; i < %v, i++ { %v } \n", stack[i-1], stack[i+1])
					}
					n, _ := stack[i-1].F.Int64()
					tmpS := make([]Var, n)
					for j := int64(0); j < n; j++ {
						tmpS[j] = stack[i+1]
					}
					tmpS2 := stack[i+2:]
					stack = append(stack[:i-1], tmpS...)
					stack = append(stack, tmpS2...)
					i -= 2
				case "depth": // Push the current stack depth
					if debug {
						fmt.Fprintf(os.Stderr, "push(len(stack))\n")
					}
					stack[i].F = big.NewFloat(float64(i))
					stack[i].Type = Number
				case "drop": // Drops the top item from the stack
					if debug {
						fmt.Fprintf(os.Stderr, "drop(stack)\n")
					}
					stack = remove(stack, i+1, 2)
					i -= 1
				case "dropn": // Drops n items from the stack
					if debug {
						fmt.Fprintf(os.Stderr, "dropn(stack,%v)\n", stack[i-1])
					}
					n, _ := stack[i-1].F.Int64()
					stack = remove(stack, i+1, int(n)+2)
					i -= int(n)
				case "dup": // Duplicates the top stack item
					if debug {
						fmt.Fprintf(os.Stderr, "x = pop(); push(%v); push(%v);\n", stack[i-1], stack[i-1])
					}
					stack[i] = stack[i-1]
				case "dupn": // Duplicates the top n stack items in order
					if debug {
						fmt.Fprintf(os.Stderr, "dupn(stack,%v)\n", stack[i-1])
					}
					n, _ := stack[i-1].F.Int64()
					tmpS := stack[i-2-int(n) : i-2]
					rest := stack[i+1:]
					stack = append(stack[:i-2], tmpS...)
					stack = append(stack, rest...)
					i -= 2
				case "roll": // Roll the stack upwards by n
					if debug {
						fmt.Fprintf(os.Stderr, "rolls the stack\n")
					}
					n, _ := stack[i-1].F.Int64()
					n = int64(int(n) % len(stack))
					tmpS := make([]Var, n)
					for j := 0; int64(j) < n; j++ {
						tmpS[j] = stack[i-2-j]
					}
					rest := stack[i+1:]
					tmpS = append(tmpS, stack[:i-int(n)-1]...)
					stack = append(tmpS, rest...)
					i -= 1
				case "rolld": // Roll the stack downwards by n
					if debug {
						fmt.Fprintf(os.Stderr, "rolls the stack by %v\n", stack[i-1])
					}
					n, _ := stack[i-1].F.Int64()
					n = int64(int(n) % len(stack))
					tmpS := make([]Var, n-1)
					for j := 0; int64(j) < n; j++ {
						tmpS[j] = stack[int(n)+j]
					}
					rest := stack[i+1:]
					tmpS = append(tmpS, stack[n-1:i]...)
					stack = append(tmpS, rest...)
					i -= 1
				case "stack": // Toggles stack display from horizontal to vertical
					if debug {
						fmt.Fprintf(os.Stderr, "toggle stack visualization %v\n", stack[i-1])
					}
					Vertical = !Vertical
				case "swap": // Swap the top 2 stack items
					if debug {
						fmt.Fprintln(os.Stderr, "Exchanging: ", stack[i-1], " and ", stack[i-2])
					}
					tmp := stack[i-2]
					stack[i-2] = stack[i-1]
					stack[i-1] = tmp
					stack = remove(stack, i+1, 1)
				case "macro": // Defines a macro, e.g. 'macro kib 1024 *'
					if debug {
						fmt.Fprintln(os.Stderr, "new macro: ", stack[i-1])
					}
					stack[i].Code = stack[i+2:]
					stack[i].Type = Code
					stack[i].V = stack[i+1].V
					vars[stack[i+1].V] = stack[i]
					keyWords[stack[i+1].V] = "x"
					if debug {
						fmt.Fprintln(os.Stderr, "Defining macro: ", stack[i+1].V, " as ", vars[stack[i+1].V].Code)
					}
					stack = remove(stack, len(stack), len(stack)-i)
					i += len(stack) - i
				case "help": // Print the help message
					flag.CommandLine.Usage()
				case "exit": // Exit the calculator
					Exit = true
					i += len(stack) - i
				default:
					if debug {
						fmt.Fprintf(os.Stderr, "Encountered user-defined macro %v -> %v\n", stack[i].V, vars[stack[i].V].Code)
					}
					rest := stack[i+1:]
					if i != 0 {
						stack = append(stack[:i], vars[stack[i].V].Code...)
					} else {
						stack = append([]Var{}, vars[stack[i].V].Code...)
					}
					i -= 1
					stack = append(stack, rest...)
				}
			} else if v, ok := vars[stack[i].V]; ok {
				// fmt.Println("asdf", v)
				stack[i] = v
				i -= 1
			} else {
				log.Fatal("TODO ", stack[i].V)
				// fmt.Println("Unknown variable: ", stack[i].V)
			}
		case Assignment:
			vars[stack[i].V] = stack[i-1]
			stack = remove(stack, i+1, 2)
		}
		if debug {
			fmt.Fprint(os.Stderr, "Evaluated as:")
			// if i < len(stack) && i >= 0 {
			fmt.Fprint(os.Stderr, stack)
			// } else {
			// 	fmt.Fprint(os.Stderr, "[] ")
			// }
			fmt.Fprintln(os.Stderr, "With variables:", vars)
		}
	}
	return stack, vars, len(stack)
}

func PrintStack(stack []Var, out *os.File) {
	format := ""
	if len(stack) > 0 {
		fmt.Fprint(out, "[ ")
	}
	for i := range stack {
		switch Mode {
		case "dec":
			if Vertical {
				format = "%v\n"
			} else {
				format = "%v,"
			}
			// fmt.Fprintf(out, format, stack[i].F)
		case "hex":
			if Vertical {
				format = "%#x\n"
			} else {
				format = "%#x,"
			}
			// fmt.Fprintf(out, format, stack[i].F)
		case "bin":
			if Vertical {
				format = "%#b\n"
			} else {
				format = "%#b,"
			}
		case "oct":
			if Vertical {
				format = "%#o\n"
			} else {
				format = "%#o,"
			}
		}
		switch stack[i].Type {
		case Number:
			if Mode == "dec" {
				fmt.Fprintf(out, format, stack[i].F)
			} else {
				tmp, _ := stack[i].F.Int(nil)
				fmt.Fprintf(out, format, tmp)
			}
		case Code:
		case Variable:
		case Assignment:
			if Vertical {
				fmt.Fprintf(out, "%v\n", stack[i].V)
			} else {
				fmt.Fprintf(out, "%v,", stack[i].V)
			}
		case String:
			fmt.Fprintf(out, format, stack[i].B)
		}
	}
	if len(stack) > 0 {
		fmt.Fprint(out, "\b ]")
	}
}

func PrintVars(vars map[string]Var, out *os.File) {
	// fmt.Fprintln(out, vars)
	if len(vars) > 0 {
		fmt.Fprintf(out, "map[ ")
	}
	format := "%v:%v,"
	for i := range vars {
		// fmt.Fprintf(out, format, i, vars[i])
		tmp, _ := vars[i].F.Int(nil)
		switch Mode {
		case "dec":
			if Vertical {
				format = "%v\n"
			} else {
				format = "%v,"
			}
			// fmt.Fprintf(out, format, stack[i].F)
		case "hex":
			if Vertical {
				format = "%#x\n"
			} else {
				format = "%#x,"
			}
			// fmt.Fprintf(out, format, stack[i].F)
		case "bin":
			if Vertical {
				format = "%#b\n"
			} else {
				format = "%#b,"
			}
		case "oct":
			if Vertical {
				format = "%#o\n"
			} else {
				format = "%#o,"
			}
		}
		switch vars[i].Type {
		case Number:
			if Mode == "dec" {
				text := vars[i].F.Text('f', int(vars[i].F.MinPrec()))
				fmt.Fprintf(out, format, text)
			} else {
				fmt.Fprintf(out, format, tmp)
			}
		case Code:
		case Variable:
		case Assignment:
			if Vertical {
				fmt.Fprintf(out, "%v\n", vars[i].V)
			} else {
				fmt.Fprintf(out, "%v,", vars[i].V)
			}
		case String:
			fmt.Fprintf(out, format, vars[i].B)
		}
	}
	if len(vars) > 0 {
		fmt.Fprint(out, "\b ]")
	}
}

func main() {
	if debug {
		fmt.Fprintln(os.Stderr, "input: ", inFile, ", output: ", outFile)
	}
	in, out := getFiles()
	defer in.Close()
	defer out.Close()
	if in == os.Stdin {
		fmt.Print("> ")
	}
	stack := make([]Var, 0)
	vars := make(map[string]Var, 0)

	inScanner := bufio.NewScanner(in)
	fp := 0
	for inScanner.Scan() {
		// out.Write([]byte( + "\n"))
		if debug {
			fmt.Fprintln(os.Stderr, "New line")
		}
		blobs := strings.Split(inScanner.Text(), " ")
		stackR, varsR := Parse(blobs)
		stack = append(stack, stackR...)
		for k := range varsR {
			vars[k] = varsR[k]
		}
		if debug {
			fmt.Fprintln(os.Stderr, "stack: ", stackR, "variables: ", varsR)
		}
		stack, vars, fp = Eval(stack, vars, fp)
		PrintStack(stack, out)
		// if debug {
		// 	PrintVars(vars, out)
		// }
		if Exit {
			os.Exit(0)
		}
		if in == os.Stdin {
			fmt.Print("> ")
		}
	}
	if err := inScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
