package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	if len(args) == 2 {
		alertInfo, err := NewSimple(args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		newFilePath := filepath.Dir(args[1]) + "/" + strings.Split(filepath.Base(args[1]), ".")[0] + "_optim" + filepath.Ext(args[1])

		f, errFile := os.Create(newFilePath)
		if errFile != nil {
			fmt.Println(errFile)
			os.Exit(2)
		}
		defer f.Close()

		alertInfo.Spool(f)
		fmt.Println("File Optimized.")
		os.Exit(0)
	}

	fmt.Println("no file provided")
}