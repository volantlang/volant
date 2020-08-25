package compiler

import (
	"error"
	"io/ioutil"
	"os"
	. "parser"
	Path "path"
	"strconv"
)

var exPath, _ = os.Executable()
var libPath = Path.Join(Path.Dir(exPath), "../lib")

var wd, _ = os.Getwd()
var DefaultC = []byte(`
int main() {
	return v0_main();
}`)

func ImportFile(dir string, base string, isMain bool, num2 int) *SymbolTable {
	n := strconv.Itoa(num)
	n2 := strconv.Itoa(num2)

	path := Path.Join(dir, base)
	OutPath := Path.Join(dir, "_build", n2+base)

	if isMain {
		OutPath += ".c"
	} else {
		OutPath += ".h"
	}

	Code, err := ioutil.ReadFile(path)

	if err != nil && !isMain {
		Code, err = ioutil.ReadFile(Path.Join(libPath, base))
	}
	if err != nil && Path.Ext(OutPath) != ".h" {
		error.NewGenError("error finding import: " + err.Error())
	}

	f, err := os.Create(OutPath)

	if err != nil {
		error.NewGenError("error creating files: " + err.Error())
	}

	if Path.Ext(path) == ".h" {
		f.Write(Code)
	} else {
		ast := ParseFile(&Lexer{Buffer: Code, Line: 1, Column: 1})
		symbols, imports, prefixes, exports, num := AnalyzeFile(ast, path)
		newAst := FormatFile(ast, symbols, imports, prefixes, num)

		if !isMain {
			f.Write([]byte("#ifndef H_" + n + "\n#define H_" + n + "\n"))
		}
		f.Write([]byte("#include \"default.h\"\n"))
		f.Write(CompileOnlyDeclarations(newAst))

		// #include \"" + OutPath + ".h\"\n
		f.Write(CompileOnlyInitializations(newAst))

		if isMain {
			f.Write(DefaultC)
		} else {
			f.Write([]byte("\n#endif"))
		}

		return exports
	}
	return &SymbolTable{}
}
