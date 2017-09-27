package helpers

import (
	"bufio"
	"fmt"
	"os"
)

// LoadOBJ - Returns vertices, UVs, and normals for a given OBJ file
func LoadOBJ(string filName) ([]float32, []float32, []float32, error) {
	var vertices, uvs, normals []float32

	objFile, err := os.Open(file)
	if err != nil {
		return vertices, uvs, normals, fmt.Errorf("obj file %q not found on disk: %v", file, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(objFile)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	return vertices, uvs, normals, nil
}
