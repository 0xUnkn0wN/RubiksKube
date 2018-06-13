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

// InitMaster -
func InitMaster() (string, error) {
	version, err := utilities.GetKubeadmVersion()
	if version == "" || err != nil {
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

		reg := regexp.MustCompile(`(?m)kubeadm join (.*)\:(.*) --token\s(.*)\s--discovery-token-ca-cert-hash sha256:(.*)`).FindAllStringSubmatch(nodeCommand, -1)

		// regex ip
		nodeParams.IP = reg[0][1]

		// regex port
		nodeParams.Port, err = strconv.Atoi(reg[0][2])
		if err != nil {
			log.Fatal(err)
		}

		// regex token
		nodeParams.Token = reg[0][3]

		// regex hash
		nodeParams.Hash = reg[0][4]

		fmt.Println("ENDPOINT: ", nodeParams.IP+":"+strconv.Itoa(nodeParams.Port))
		fmt.Println("TOKEN: ", nodeParams.Token)
		fmt.Println("PORT: ", nodeParams.Port)
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

	return "", err
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
