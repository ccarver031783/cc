package eksupgrade

import (
	"os"
	"path/filepath"
	"strings"
)

// DiscoverClusters finds all EKS cluster directories under infraRoot.
// A cluster dir is one containing terraform.tfvars and eks.tf.
func DiscoverClusters(infraRoot string) ([]string, error) {
	var clusters []string
	err := filepath.Walk(infraRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		// Check if this directory has terraform.tfvars and eks.tf
		tfvars := filepath.Join(path, "terraform.tfvars")
		eks := filepath.Join(path, "eks.tf")
		if fileExists(tfvars) && fileExists(eks) {
			rel, err := filepath.Rel(infraRoot, path)
			if err != nil {
				return err
			}
			// Only include paths under aws/ (EKS clusters)
			if strings.HasPrefix(rel, "aws"+string(filepath.Separator)) {
				clusters = append(clusters, rel)
			}
		}
		return nil
	})
	return clusters, err
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
