package model

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

/* has everything */
type Policy struct {
	PolicyDescription string //one user one policy or multiple?
	PolicyTreeRoot    *PolicyNode
	AccessStruct      *AccessStruct
}

func (p *Policy) Grow() *Policy {
	if p.PolicyDescription == "" {
		fmt.Printf("error:: user's policy description is EMPTY.\n")
		return nil
	} else {
		policyStr := p.PolicyDescription
		p.PolicyTreeRoot, p.AccessStruct = ParsePolicyStringToTree(&policyStr)
		//fmt.Println(p.AccessStruct)
		p.AccessStruct.genLsssMatrix()
		p.AccessStruct.padLsssMatrix()
		p.AccessStruct.genPolicyTreePath()
	}
	return p
}

/*  Policy Node */
type PolicyNode struct {
	Type      byte //0: node. 1: leaf
	Operation byte //1: and. 2: or
	Attr      string
	Max       int
	Min       int
	Children  []*PolicyNode
}

func NewPolicyNode(attr string, t byte) *PolicyNode {
	N := new(PolicyNode)
	N.Attr = attr
	N.Type = t
	N.Operation = 0
	N.Max = 0
	N.Min = 0
	N.Children = nil
	//fmt.Printf("NewPolicyNode:: %v\n",Attr)
	return N
}
func (n *PolicyNode) GetAttr() string {
	return n.Attr
}
func (n *PolicyNode) SetAttr(attr string) *PolicyNode {
	n.Attr = attr
	return n
}
func (n *PolicyNode) GetMax() int {
	return n.Max
}
func (n *PolicyNode) SetMax(max int) *PolicyNode {
	n.Max = max
	return n
}
func (n *PolicyNode) GetOperation() byte {
	return n.Operation
}
func (n *PolicyNode) SetOperation(o byte) *PolicyNode {
	n.Operation = o
	return n
}
func (n *PolicyNode) GetMin() int {
	return n.Min
}
func (n *PolicyNode) SetMin(min int) *PolicyNode {
	n.Min = min
	return n
}
func (n *PolicyNode) GetChildren() []*PolicyNode {
	return n.Children
}
func (n *PolicyNode) SetChildren(children []*PolicyNode) *PolicyNode {
	n.Children = children
	return n
}

type sendpon struct {
	Type      byte //0: node. 1: leaf
	Operation byte //1: and. 2: or
	Attr      string
	Max       int
	Min       int
	Children  [][]byte
}

func (b *PolicyNode) Serialize() []byte {
	var result bytes.Buffer
	var spon *sendpon = new(sendpon)
	spon.Type = b.Type
	spon.Operation = b.Operation
	spon.Attr = b.Attr
	spon.Max = b.Max
	spon.Min = b.Min
	for _, c := range b.Children {
		spon.Children = append(spon.Children, c.Serialize())
	}
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(spon)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializePON(d []byte) *PolicyNode {
	pon := new(PolicyNode)
	var spon sendpon

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&spon)
	if err != nil {
		log.Panic(err)
	}

	pon.Type = spon.Type
	pon.Operation = spon.Operation
	pon.Attr = spon.Attr
	pon.Max = spon.Max
	pon.Min = spon.Min
	for _, c := range spon.Children {
		pon.Children = append(pon.Children, DeserializePON(c))
	}

	return pon
}

type sendpy struct {
	PolicyDescription string //one user one policy or multiple?
	PolicyTreeRoot    []byte
	AccessStruct      []byte
}

func (b *Policy) Serialize() []byte {
	var result bytes.Buffer
	var spy *sendpy = new(sendpy)
	spy.PolicyDescription = b.PolicyDescription
	spy.PolicyTreeRoot = b.PolicyTreeRoot.Serialize()
	spy.AccessStruct = b.AccessStruct.Serialize()

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(spy)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializePY(d []byte, cp *CurveParam) *Policy {
	py := new(Policy)
	var spy sendpy

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&spy)
	if err != nil {
		log.Panic(err)
	}

	py.PolicyDescription = spy.PolicyDescription
	py.PolicyTreeRoot = DeserializePON(spy.PolicyTreeRoot)
	py.AccessStruct = DeserializeACS(spy.AccessStruct, cp)

	return py
}
