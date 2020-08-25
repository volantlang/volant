package main

import (
	. "compiler"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

var exPath, _ = os.Executable()
var libPath = path.Join(path.Dir(exPath), "../lib")
var defaultC, _ = ioutil.ReadFile(path.Join(libPath, "internal/default.h"))

func main() {
	fileName := flag.String("compile", "", "file to be compiled")
	flag.Parse()

	file := path.Clean(*fileName)

	if file == "" {
		fmt.Println("file name not given")
		os.Exit(1)
	}

	fd, _ := os.Create(path.Join(path.Dir(file), "_build/default.h"))
	fd.Write(defaultC)

	ImportFile(path.Dir(file), path.Base(file), true, 0)
	out, err := exec.Command("clang", path.Join(path.Dir(file), "_build", "0"+path.Base(file)+".c"), "-pthread", "-luv", "-fblocks", "-lBlocksRuntime", "-lgc", "-o", path.Join(path.Dir(file), "a.out")).CombinedOutput()

	if err != nil {
		fmt.Println(string(out))
	}
}
