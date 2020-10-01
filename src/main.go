package main

import (
	. "compiler"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
)

var exPath, _ = os.Executable()
var libPath = path.Join(path.Dir(exPath), "../lib")
var defaultH = path.Join(libPath, "internal/default.h")

func subcommands() { // Helper function
	fmt.Println("  compile string")
	fmt.Println("    compile a file")
	fmt.Println("Valid flags are:")
}

func main() {
	compile := flag.NewFlagSet("compile", flag.ExitOnError)
	clang := flag.String("clang", "", "pass arguments to clang")

	if len(os.Args) < 2 {
		fmt.Println("\x1b[31mError: a subcommand is required\x1b[0m")
		fmt.Println("Valid subcommands are:")
		subcommands()
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "compile":
		compile.Parse(os.Args[2:])
	default:
		fmt.Println("\x1b[31mError: invalid subcommand\x1b[0m")
		subcommands()
		fmt.Println("Valid flags are:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if compile.Parsed() {
		file := compile.Arg(0)

		if file == "" {
			fmt.Println("\x1b[31mError: a file is required for `compile`\x1b[0m")
			os.Exit(1)
		}

		file = path.Clean(file)

		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Println(fmt.Sprintf("\x1b[31mError: file `%s` does not exist\x1b[0m", file))
			os.Exit(1)
		}

		ImportFile(path.Dir(file), path.Base(file), true, 0)

		out, err := exec.Command("clang", path.Join(path.Dir(file), "_build", "0"+path.Base(file)+".c"), "-pthread", "-luv", "-fblocks", "-lBlocksRuntime", "-lgc", "-I"+libPath, *clang, "-o", path.Join(path.Dir(file), "a.out")).CombinedOutput()

		if err != nil {
			fmt.Println(string(out))
			os.Exit(1)
		}
	}
}
