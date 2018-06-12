package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/JonathanHeinz/RubiksKube/master"
	"github.com/JonathanHeinz/RubiksKube/node"
	"github.com/JonathanHeinz/RubiksKube/utilities"
)

var nodeHash = ""
var createMaster = false

func main() {
	// flags
	flag.StringVar(&nodeHash, "add-node", "", "Add a node")
	flag.BoolVar(&createMaster, "init-master", false, "Create a master")
	flag.Parse()

	// is ubuntu
	isNotUbuntu, err := utilities.IsUbuntu()
	if !isNotUbuntu || err != nil {
		log.Fatal("Not a ubuntu operation system")
	}
	// check swap off
	swapNotEnabled, err := utilities.SwapEnabled()
	if swapNotEnabled || err != nil {
		log.Fatal("Running with swap on is not supported, please disable swap!")
	}

	if err := utilities.InitDocker(); err != nil {
		log.Fatal("docker: not successfully initialized")
		log.Fatal(err)
		return
	}
	fmt.Println("docker: ready")

	if err := utilities.InitKubeadm(); err != nil {
		log.Fatal("kubeadm: not successfully initialized")
		log.Fatal(err)
		return
	}
	fmt.Println("kubeadm: ready")

	if createMaster {
		// init as master
		hash, err := master.InitMaster()
		if err != nil {
			log.Fatal("master creation failed")
		}

		fmt.Printf("run 'rubikskube --add-node=%s' to add another server to the cluster", hash)
	}

	if nodeHash != "" && !createMaster {
		if node.InitNode(nodeHash) != nil {
			log.Fatal("node initial failed")
		}
	}
}
