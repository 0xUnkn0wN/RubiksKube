package utilities

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// PodNetwork -
type PodNetwork struct {
	RBACURL   string
	URL       string
	Definiton string
}

// InitKubernetes -
// func InitKubernetes() {
// 	initKubeadm()
// 	initDocker()
// }

// SwapEnabled -
func SwapEnabled() (bool, error) {
	// Check whether swap is enabled. The Kubelet does not support running with swap enabled.
	swapData, err := ioutil.ReadFile("/proc/swaps")
	if err != nil {
		return false, err
	}

	swapData = bytes.TrimSpace(swapData) // extra trailing \n
	swapLines := strings.Split(string(swapData), "\n")

	// If there is more than one line (table headers) in /proc/swaps, swap is enabled
	if len(swapLines) > 1 {
		return true, nil
	}

	return false, nil
}

// IsUbuntu -
func IsUbuntu() (bool, error) {
	osInfo, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return false, err
	}

	return regexp.MustCompile("(?i)ubuntu").MatchString(string(osInfo)), nil
}

// GetCanalNetwork -
func GetCanalNetwork() PodNetwork {
	return PodNetwork{
		RBACURL:   "https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/rbac.yaml",
		URL:       "https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/canal.yaml",
		Definiton: "Canal",
	}
}

// CP -
func CP(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		return err
	}

	return dstFile.Close()
}
