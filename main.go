package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
)

var masterPort = 6443
var masterIP = ""
var nodeHash string

type podNetwork struct {
	RBACURL   string
	URL       string
	Definiton string
}

type params struct {
	IP    string
	Port  int
	Hash  string
	Token string
}

func main() {
	// flags
	flag.StringVar(&nodeHash, "add-node", "", "")

	// is ubuntu
	isNotUbuntu, err := isUbuntu()
	if !isNotUbuntu || err != nil {
		log.Fatal("Not a ubuntu operation system")
	}
	// check swap off
	swapNotEnabled, err := swapEnabled()
	if swapNotEnabled || err != nil {
		log.Fatal("Running with swap on is not supported, please disable swap!")
	}
	// add kubeadm to source list

	// install docker TODO: check if docker is installed
	if initDocker() != nil {
		log.Fatal("docker not successfully initialized")
	}
	fmt.Println("Docker status: ready")

	// install kubeadm TODO: check if kubeadm is installed
	if initKubeadm() != nil {
		log.Fatal("kubeadmin not successfully initialized")
	}
	fmt.Println("kubeadm: ready")

	if nodeHash != "" {
		// init as master
		hash, err := initMaster()
		if err != nil {
			log.Fatal("master creation failed")
		}

		fmt.Printf("run 'rubikskube --add-node=%s' to add another server to the cluster", hash)
		return
	}

	err = initNode()
	if err != nil {
		log.Fatal("node initial failed")
	}
}

func createClusterUser() error {
	currentUser, err := user.Current()
	if err != nil {
		return err
	}

	kubePath := currentUser.HomeDir + "/.kube"

	err = os.MkdirAll(kubePath, 0666)
	if err != nil {
		return err
	}

	cp(kubePath+"/config", "/etc/kubernetes/admin.conf")

	gid, err := strconv.ParseInt(currentUser.Gid, 0, 0)
	if err != nil {
		return err
	}

	uid, err := strconv.ParseInt(currentUser.Uid, 0, 0)
	if err != nil {
		return err
	}

	err = os.Chown(kubePath+"/config", int(uid), int(gid))
	if err != nil {
		return err
	}

	return nil
}

func applyToCluster(filename string) error {
	return exec.Command("kubectl", "apply", "-f", filename).Run()
}

func initMaster() (string, error) {
	var nodeCommandAsBytes []byte
	var nodeParams params
	// kubeadm init
	response, err := exec.Command("kubeadm", "init").Output()
	if err != nil {
		return "", err
	}

	index := strings.LastIndex(string(response), "kubeadm join")
	_, err = strings.NewReader(string(response)).ReadAt(nodeCommandAsBytes, int64(index))
	if err != nil {
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
	err = createClusterUser()
	if err != nil {
		return "", err
	}

	// init network pod
	networkService := getCanalNetwork()

	err = applyToCluster(networkService.URL)
	if err != nil {
		return "", err
	}

	err = applyToCluster(networkService.RBACURL)
	if err != nil {
		return "", err
	}

	// -> init master -> return base64 (token_ip:port_hash) -> (maybe) save command
	paramsAsJSON, err := json.Marshal(nodeParams)
	if err != nil {
		log.Fatal("hash parsing error")
	}

	return base64.RawStdEncoding.EncodeToString(paramsAsJSON), nil
}

func initNode() error {
	var nodeParams params
	// -> join node -> decode base64 -> execute add command
	paramsAsJSON, err := base64.RawStdEncoding.DecodeString(nodeHash)
	if err != nil {
		return err
	}

	err = json.Unmarshal(paramsAsJSON, nodeParams)
	if err != nil {
		return err
	}

	// kubeadm join --token <token> <master-ip>:<master-port> --discovery-token-ca-cert-hash sha256:<hash>
	err = exec.Command("kubeadm", "join", "--token", nodeParams.Token, nodeParams.IP+":"+strconv.Itoa(nodeParams.Port), "--discovery-token-ca-cert-hash shar256:"+nodeParams.Hash).Run()
	if err != nil {
		return err
	}

	return nil
}

func initDocker() error {
	// TODO: check if docker is already installed and maybe check the right version
	err := exec.Command("apt-get", "update").Run()
	if err != nil {
		return err
	}

	err = exec.Command("apt-get", "install", "-y", "docker.io").Run()
	if err != nil {
		return err
	}

	return nil
}

func initKubeadm() error {
	// TODO: check if kubeadm is already installed and maybe check the right version
	resp, err := http.Get("https://packages.cloud.google.com/apt/doc/apt-key.gpg")
	if err != nil {
		return err
	}

	responseContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	addSourceCommand := exec.Command("add-key", "-")
	addSourceCommandPipe, err := addSourceCommand.StdinPipe()
	if err != nil {
		return err
	}

	err = addSourceCommand.Start()
	if err != nil {
		return err
	}

	_, err = addSourceCommandPipe.Write(responseContent)
	if err != nil {
		return err
	}

	sourceFile, err := os.Create("/etc/apt/sources.list.d/kubernetes.list")
	if err != nil {
		return err
	}

	_, err = sourceFile.WriteString("deb http://apt.kubernetes.io/ kubernetes-xenial main")
	if err != nil {
		return err
	}

	err = exec.Command("apt-get", "update").Run()
	if err != nil {
		return err
	}

	err = exec.Command("apt-get", "install", "-y", "kubelet kubeadm kubectl").Run()
	if err != nil {
		return err
	}

	return nil
}

func swapEnabled() (bool, error) {
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

func isUbuntu() (bool, error) {
	osInfo, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return false, err
	}

	return regexp.MustCompile("(?i)ubuntu").MatchString(string(osInfo)), nil
}

func getCanalNetwork() podNetwork {
	return podNetwork{
		RBACURL:   "https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/rbac.yaml",
		URL:       "https://raw.githubusercontent.com/projectcalico/canal/master/k8s-install/1.7/canal.yaml",
		Definiton: "Canal",
	}
}

func cp(dst, src string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		dstFile.Close()
		return err
	}

	return dstFile.Close()
}
