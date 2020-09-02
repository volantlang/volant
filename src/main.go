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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no arguments given")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "compile":
		if len(os.Args) < 3 {
			fmt.Println("file name not given")
		}

		cmd := flag.NewFlagSet("compile", flag.ExitOnError)
		clang := cmd.String("clang", "", "pass arguments to the clang compiler")

		file := path.Clean(os.Args[2])
		cmd.Parse(os.Args[3:])

		ImportFile(path.Dir(file), path.Base(file), true, 0)

		out, err := exec.Command("/bin/bash", "-c", "clang " + path.Join(path.Dir(file), "_build", "0"+path.Base(file)+".c") + " -pthread " + "-luv" + " -fblocks " + " -lBlocksRuntime " + " -lgc " + " -I" + libPath + " " + *clang + " -o " + path.Join(path.Dir(file), "a.out")).CombinedOutput()
		
		if err != nil {
			fmt.Println(string(out))
			os.Exit(1)
		}
	}
}
