package main

import "crypto/sha256"

type MerkelNode struct {
	Left  *MerkelNode
	Right *MerkelNode
	Data  []byte
}

type MerkelTree struct {
	Root *MerkelNode
}

func NewMerkelNode(left, right *MerkelNode, data []byte) *MerkelNode {
	newNode := MerkelNode{}
	// merkel tree is a complete two-branch tree
	if left != nil && right != nil {
		data = append(left.Data, right.Data...) // not null and take their data
	}
	hashs := sha256.Sum256(data) //  if left and right == nil ,data is from the outside data (usually transaction)
	newNode.Data = hashs[:]
	newNode.Left = left
	newNode.Right = right
	return &newNode
}

func NewMerkelTree(data [][]byte) MerkelTree {
	var nodes []MerkelNode
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	for _, d := range data {
		newNode := NewMerkelNode(nil, nil, d)
		nodes = append(nodes, *newNode)
	}
Tree:
	for {
		var upperLevel []MerkelNode
		for j := 0; j < len(nodes); j += 2 {
			newNode := NewMerkelNode(&nodes[j], &nodes[j+1], nil)
			upperLevel = append(upperLevel, *newNode)
		}
		nodes = upperLevel
		if len(nodes) == 1 {
			break Tree
		}
	}
	return MerkelTree{&nodes[0]}
}
