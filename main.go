package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	args := os.Args

	var err error
	var alertInfo *AlertInfo
	var fileName string

	if len(args) == 2 {
		fileName = args[1]
		alertInfo, err = NewSimple(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fileName = "sample.yaml"
		alertInfo, err = NewSimple(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	newFilePath := filepath.Dir(fileName) + "/" + strings.Split(filepath.Base(fileName), ".")[0] + "_optim" + filepath.Ext(fileName)

	f, errFile := os.Create(newFilePath)
	if errFile != nil {
		fmt.Println(errFile)
		os.Exit(2)
	}
	defer f.Close()

	alertInfo.Spool(f)

	fmt.Println("File Optimized.")
	fmt.Println("ZUpper: ", len(alertInfo.AlertsZUpper))
	fmt.Println("ZLower: ", len(alertInfo.AlertsZLower))
	fmt.Println("LUpper: ", len(alertInfo.AlertsLUpper))
	fmt.Println("LLower: ", len(alertInfo.AlertsLLower))
	fmt.Println("Other:  ", len(alertInfo.AlertsOther))

	os.Exit(0)

	//	fmt.Println("no file provided")
}
