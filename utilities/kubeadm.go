package utilities

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
)

// InitKubeadm -
func InitKubeadm() error {
	version, err := GetKubeadmVersion()
	if version == "" || err != nil {
		resp, err := http.Get("https://packages.cloud.google.com/apt/doc/apt-key.gpg")
		if err != nil {
			return err
		}

		responseContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		tmp, err := ioutil.TempFile("", "key")
		if err != nil {
			return err
		}

		if _, err := tmp.Write(responseContent); err != nil {
			return err
		}

		if err := tmp.Close(); err != nil {
			return err
		}

		if err := exec.Command("apt-key", "add", tmp.Name()).Run(); err != nil {
			return err
		}
		os.Remove(tmp.Name())

		sourceFile, err := os.Create("/etc/apt/sources.list.d/kubernetes.list")
		if err != nil {
			return err
		}

		if _, err := sourceFile.WriteString("deb http://apt.kubernetes.io/ kubernetes-xenial main"); err != nil {
			return err
		}

		if err := exec.Command("apt-get", "update").Run(); err != nil {
			return err
		}

		if err := exec.Command("apt-get", "install", "-y", "kubelet", "kubeadm", "kubectl").Run(); err != nil {
			return err
		}
	}

	return nil
}

// GetKubeadmVersion -
func GetKubeadmVersion() (string, error) {
	// docker version --format
	version, err := exec.Command("kubeadm", "version", "-o", "short").Output()
	if err != nil {
		return "", err
	}

	return string(version), nil
}
