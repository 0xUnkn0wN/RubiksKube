package node

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"strconv"
)

// NodeParams -
type NodeParams struct {
	IP    string
	Port  int
	Hash  string
	Token string
}

// InitNode -
func InitNode(hash string) error {
	var params NodeParams
	// -> join node -> decode base64 -> execute add command
	paramsAsJSON, err := base64.RawStdEncoding.DecodeString(hash)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(paramsAsJSON, params); err != nil {
		return err
	}

	// kubeadm join --token <token> <master-ip>:<master-port> --discovery-token-ca-cert-hash sha256:<hash>
	if err := exec.Command("kubeadm", "join", "--token", params.Token, params.IP+":"+strconv.Itoa(params.Port), "--discovery-token-ca-cert-hash shar256:"+params.Hash).Run(); err != nil {
		return err
	}

	return nil
}
