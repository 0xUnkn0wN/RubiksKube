package utilities

import (
	"os/exec"
)

// InitDocker -
func InitDocker() error {
	version, err := GetDockerVersion()
	if err != nil {
		return err
	}

	if version == "" {
		if err := exec.Command("apt-get", "update").Run(); err != nil {
			return err
		}

		if err := exec.Command("apt-get", "install", "-y", "docker.io").Run(); err != nil {
			return err
		}
	}

	return nil
}

// GetDockerVersion -
func GetDockerVersion() (string, error) {
	// docker version --format
	version, err := exec.Command("docker", "version", "--format", "'{{.Client.Version}}'").Output()
	if err != nil {
		return "", err
	}

	return string(version), nil
}
