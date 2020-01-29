package MulticenterABEForFabric

import (
	"fmt"
	"github.com/Nik-U/pbc"
	"github.com/cloudflare/cfssl/scan/crypto/sha256"
)

type Authority struct {
	PKMap    map[string]*PK
	SKMap    map[string]*SK
	EGGAlpha *pbc.Element
	Alpha    *pbc.Element
	Name     string
}

func (a *Authority) GenerateNewAttr(attr string, curve *CurveParam, g *pbc.Element) (*PK, error) {
	if a.PKMap[attr] != nil || a.SKMap[attr] != nil {
		return nil, fmt.Errorf("already has this attr:%s", attr)
	}
	y := curve.GetNewZn()
	sk := SK{y}
	gy := g.NewFieldElement().PowZn(g, y)
	pk := PK{gy}
	a.PKMap[attr] = &pk
	a.SKMap[attr] = &sk
	return &pk, nil
}

func (a *Authority) KeyGen(gid string, g *pbc.Element, attr string) (*pbc.Element, error) {
	if a.SKMap[attr] == nil {
		return nil, fmt.Errorf("don't have this attr, error when %s", attr)
	}
	temp := sha256.Sum256([]byte(gid))
	hashGid := g.NewFieldElement().SetBytes(temp[:])
	key := g.NewFieldElement().PowZn(g, a.Alpha).ThenMul(hashGid.ThenPowZn(a.SKMap[attr].Y))
	return key, nil
}