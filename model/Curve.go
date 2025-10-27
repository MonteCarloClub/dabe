package model

import (
	"hash"
	"math/big"

	"github.com/Nik-U/pbc"
)

type CurveParam struct {
	p *big.Int //Order of G, N=p1*p2*p3

	Param   *pbc.Params
	Pairing *pbc.Pairing
}

func (cp *CurveParam) Initialize() {
	cp.p = new(big.Int)
	p1 := new(big.Int)
	p2 := new(big.Int)
	p3 := new(big.Int)
	p1.SetString("242661090146032969904098483991985908921", 10) // octal
	p2.SetString("215662396313044988944834777682074105079", 10) // octal
	p3.SetString("253493408475411572624002367871313476827", 10) // octal
	cp.p.Mul(p1, p2)
	cp.p.Mul(cp.p, p3)
	cp.Param = pbc.GenerateA1(cp.p)
	cp.Pairing = cp.Param.NewPairing()
}

func (cp *CurveParam) GetP() *big.Int {
	N := new(big.Int)
	N.Set(cp.p)
	return N
}
func (cp *CurveParam) GetPairing() *pbc.Pairing {
	return cp.Pairing
}

func (cp *CurveParam) GetNewG1() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(0).Rand()
	return g
}

func (cp *CurveParam) GetNewGT() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(2).Rand()
	return g
}

func (cp *CurveParam) GetNewZn() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(3).Rand()
	return g
}

func (cp *CurveParam) GetG1FromStringHash(s string, hash hash.Hash) *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(0).SetFromStringHash(s, hash)
	return g
}

func (cp *CurveParam) GetZnFromStringHash(s string, hash hash.Hash) *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(3).SetFromStringHash(s, hash)
	return g
}

func (cp *CurveParam) Get0FromG1() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(0).Set0()
	return g
}

func (cp *CurveParam) Get0FromGT() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(2).Set0()
	return g
}

func (cp *CurveParam) Get0FromZn() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(3).Set0()
	return g
}

func (cp *CurveParam) Get1FromG1() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(0).Set1()
	return g
}

func (cp *CurveParam) Get1FromGT() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(2).Set1()
	return g
}

func (cp *CurveParam) Get1FromZn() *pbc.Element {
	g := cp.Pairing.NewUncheckedElement(3).Set1()
	return g
}
