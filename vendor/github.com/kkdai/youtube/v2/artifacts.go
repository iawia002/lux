package youtube

import (
	"log"
	"os"
	"path/filepath"
)

// destination for artifacts, used by integration tests
var artifactsFolder = os.Getenv("ARTIFACTS")

func writeArtifact(name string, content []byte) {
	path := filepath.Join(artifactsFolder, name)
	err := os.WriteFile(path, content, 0600)
	if err != nil {
		log.Printf("unable to write artifact %s: %v", path, err)
	} else {
		log.Println("artifact created:", path)
	}
}
