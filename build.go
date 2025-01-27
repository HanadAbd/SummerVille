package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const srcDir = "web/src"
const distDir = "web/dist"

func doBuild() {
	err := os.RemoveAll(distDir)
	if err != nil {
		fmt.Printf("Error removing dist directory: %v\n", err)
		return
	}

	err = os.MkdirAll(distDir, os.ModePerm)
	if err != nil {
		fmt.Printf("Error creating dist directory: %v\n", err)
		return
	}

	err = copyAndCompile(srcDir, distDir)
	if err != nil {
		fmt.Printf("Error copying and compiling: %v\n", err)
		return
	}

	fmt.Println("Files copied and compiled successfully.")

	// cmd := exec.Command("npx", "webpack")
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Printf("Error running Webpack: %v, output: %s\n", err, string(output))
	// 	return
	// }

	// fmt.Println("Webpack compilation completed successfully.")
}

func copyAndCompile(src, dist string) error {
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return fmt.Errorf("error reading source directory: %v", err)
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		distPath := filepath.Join(dist, file.Name())

		if file.IsDir() {
			err := os.MkdirAll(distPath, os.ModePerm)
			if err != nil {
				return fmt.Errorf("error creating directory: %v", err)
			}
			err = copyAndCompile(srcPath, distPath)
			if err != nil {
				return err
			}
		} else if filepath.Ext(file.Name()) != ".ts" {
			input, err := ioutil.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("error reading file: %v", err)
			}
			err = ioutil.WriteFile(distPath, input, file.Mode())
			if err != nil {
				return fmt.Errorf("error writing file: %v", err)
			}
		}
	}

	return nil
}
