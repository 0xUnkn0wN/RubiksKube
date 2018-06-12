package master

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"

	"github.com/JonathanHeinz/RubiksKube/node"
	"github.com/JonathanHeinz/RubiksKube/utilities"
)

// DefaultPort -
var DefaultPort = 6443

// InitMaster -
func InitMaster() (string, error) {
	var nodeCommandAsBytes []byte
	var nodeParams node.NodeParams
	// kubeadm init
	response, err := exec.Command("kubeadm", "init").Output()
	if err != nil {
		return "", err
	}

	index := strings.LastIndex(string(response), "kubeadm join")

	if _, err := strings.NewReader(string(response)).ReadAt(nodeCommandAsBytes, int64(index)); err != nil {
		return "", err
	}

	nodeCommand := string(nodeCommandAsBytes)

	// regex token
	nodeParams.Token = regexp.MustCompile(`--token\s(.*)\s-`).FindAllStringSubmatch(nodeCommand, 0)[0][0]

	fmt.Println("TOKEN: ", nodeParams.Token)
	// regex ip
	nodeParams.IP = regexp.MustCompile(`kubeadm\sjoin\s([0-9\.]*)`).FindAllStringSubmatch(nodeCommand, 0)[0][0]

	fmt.Println("IP: ", nodeParams.IP)
	// regex port
	nodeParams.Port, err = strconv.Atoi(regexp.MustCompile(`kubeadm\sjoin.*\:([0-9]*)\s`).FindAllStringSubmatch(nodeCommand, 0)[0][0])
	if err != nil {
		log.Fatal("can't convert to int")
	}

	fmt.Println("PORT: ", nodeParams.Port)
	// regex hash
	nodeParams.Hash = regexp.MustCompile(`sha256:(.*)`).FindAllStringSubmatch(nodeCommand, 0)[0][0]

	fmt.Println("HASH: ", nodeParams.Hash)

	// create cluster user
	if err := createClusterUser(); err != nil {
		return "", err
	}

	// init network pod
	networkService := utilities.GetCanalNetwork()

	if err := applyToCluster(networkService.URL); err != nil {
		return "", err
	}

	if err := applyToCluster(networkService.RBACURL); err != nil {
		return "", err
	}

	// -> init master -> return base64 (token_ip:port_hash) -> (maybe) save command
	paramsAsJSON, err := json.Marshal(nodeParams)
	if err != nil {
		log.Fatal("hash parsing error")
	}

	return base64.RawStdEncoding.EncodeToString(paramsAsJSON), nil
}

func createClusterUser() error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	kubePath := currentUser.HomeDir + "/.kube"

	if err := os.MkdirAll(kubePath, 0666); err != nil {
		return err
	}

	utilities.CP(kubePath+"/config", "/etc/kubernetes/admin.conf")

	gid, err := strconv.ParseInt(currentUser.Gid, 0, 0)
	if err != nil {
		return err
	}

	uid, err := strconv.ParseInt(currentUser.Uid, 0, 0)
	if err != nil {
		return err
	}

	if err := os.Chown(kubePath+"/config", int(uid), int(gid)); err != nil {
		return err
	}

	return nil
}

func applyToCluster(filename string) error {
	return exec.Command("kubectl", "apply", "-f", filename).Run()
}
