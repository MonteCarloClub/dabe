package DecentralizedABE

import (
	"github.com/Nik-U/pbc"
)

/* Public Key Structure */
type APK struct {
	Gy *pbc.Element //G^y, y from Zp
}

func (p *APK) Initialize(gy *pbc.Element) {
	p.Gy = gy
}

func (p *APK) getGy() *pbc.Element {
	return p.Gy.NewFieldElement().Set(p.Gy)
}

type ASK struct {
	Y *pbc.Element
}

func (s *ASK) Initialize(y *pbc.Element) {
	s.Y = y
}

func (s *ASK) getY() *pbc.Element {
	return s.Y.NewFieldElement().Set(s.Y)
}

type OPKPart struct {
	EGGAlphaPart *pbc.Element            //part of org's EGGAlpha
	GyPart       map[string]*pbc.Element //part of org attrs' gy
}
type OSKPart struct {
	AlphaPart   *pbc.Element        //part of org's Alpha
	ASKPartMap  map[string]*ASKPart //part of org attrs' y
	F           []*pbc.Element      //for shamir's share
	N           int
	T           int
	OthersShare []*pbc.Element //for some special time
}
type ASKPart struct {
	F           []*pbc.Element //for shamir's share
	OthersShare []*pbc.Element //for some special time
	YPart       *pbc.Element
}
