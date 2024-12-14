package nodes

import (
	"encoding/json"
	"fmt"
)

// go:embed nodes.json
var nodesConfig []byte

type Node struct {
	Name      string               `json:"name"`
	Prev      *Node                `json:"prev"`
	Run       func([]string) error `json:"-"`
	Config    *Config              `json:"config"`
	BranchMap map[string]*Node     `json:"branches"`
}

type Config struct {
	Package  string `json:"package"`
	File     string `json:"file"`
	Function string `json:"function"`
	Args     *[]struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Required bool   `json:"required"`
	} `json:"args,omitempty"`
}

func NewNode(name string, prev *Node, config *Config) *Node {
	node := &Node{
		Name:      name,
		Prev:      prev,
		Config:    config,
		BranchMap: make(map[string]*Node),
	}

	if prev != nil {
		prev.BranchMap[name] = node
	}

	return node
}

func LoadNodes() (*Node, error) {
	// unmarshal directly from embedded data
	var temp struct {
		Name      string  `json:"name"` // root.Name == 'swan'
		Config    *Config `json:"config"`
		BranchMap map[string]*struct {
			Name      string           `json:"name"`
			Config    *Config          `json:"config"`
			BranchMap map[string]*Node `json:"branches"`
		} `json:"branches"`
	}

	if err := json.Unmarshal(nodesConfig, &temp); err != nil {
		return nil, fmt.Errorf("error parsing config: %v", err)
	}

	// create swan node
	root := NewNode(temp.Name, nil, temp.Config)

	// create branch nodes for swan
	for branchName, branchData := range temp.BranchMap {
		if branchName != branchData.Name {
			return nil, fmt.Errorf("branch name mismatch: %s != %s",
				branchName, branchData.Name)
		}
		NewNode(branchData.Name, root, branchData.Config)
	}

	// swan
	return root, nil
}
