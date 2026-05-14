// Package main is the entry point for the expense tracker API server.
//
// In Go, the "main" package is special: it's the only package that produces
// an executable binary when compiled. Every other package compiles into a
// library that other code imports.
//
// This file's only job (for now) is to print "Hello, expense tracker!" so we
// can confirm our toolchain is wired up correctly. We'll grow it into the
// real server one piece at a time.
package main

// "fmt" is Go's standard formatted-I/O package — it's how you print to stdout,
// format strings, etc. Roughly analogous to Python's print() / f-strings.
//
// Imports in Go are ALWAYS explicit: you must import every package you use,
// and you can't import something you don't use (the compiler will refuse to
// build). This forces tidy code.
import "fmt"

// main is the function Go calls when you run a program in the "main" package.
// It takes no arguments and returns nothing. The signature is fixed.
//
// (If you've used Python: think `if __name__ == "__main__":`. If C/C++: same
// as `int main()` but simpler — no exit code, no argc/argv. Use os.Args if you
// want command-line arguments later.)
func main() {
	// Println adds a newline at the end. There's also Print (no newline)
	// and Printf (formatted, like C's printf).
	fmt.Println("Hello, expense tracker!")
}
