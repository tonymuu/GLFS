package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	fmt.Println(exPath)
}
