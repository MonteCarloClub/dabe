package DecentralizedABE

import (
	"fmt"
	"github.com/Nik-U/pbc"
	"github.com/cloudflare/cfssl/scan/crypto/sha256"
)

type User struct {
	APKMap   map[string]*APK
	ASKMap   map[string]*ASK
	EGGAlpha *pbc.Element
	Alpha    *pbc.Element
	Name     string
	OPKMap	map[string]*OPKPart
	OSKMap	map[string]*OSKPart
}

func (u *User) GetPK() *pbc.Element {
	return u.EGGAlpha
}

func (u *User) GetAPKMap() map[string]*APK {
	return u.APKMap
}

func (u *User) GenerateNewAttr(attr string, d *DABE) (*APK, error) {
	if u.APKMap[attr] != nil || u.ASKMap[attr] != nil {
		return nil, fmt.Errorf("already has this attr:%s", attr)
	}
	y := d.CurveParam.GetNewZn()
	sk := ASK{y}
	gy := d.G.NewFieldElement().PowZn(d.G, y)
	pk := APK{gy}
	u.APKMap[attr] = &pk
	u.ASKMap[attr] = &sk
	return &pk, nil
}

func (u *User) KeyGen(gid string, attr string, d *DABE) (*pbc.Element, error) {
	if u.ASKMap[attr] == nil {
		return nil, fmt.Errorf("don't have this attr, error when %s", attr)
	}
	temp := sha256.Sum256([]byte(gid))
	hashGid := d.G.NewFieldElement().SetBytes(temp[:])
	key := d.G.NewFieldElement().PowZn(d.G, u.Alpha).ThenMul(hashGid.ThenPowZn(u.ASKMap[attr].Y))
	return key, nil
}

func (u *User) GenerateOrgShare(n,t int, userNames map[string]*pbc.Element) ([]*pbc.Element, error) {

}

//get sij
func (u *User) share(idb []uint8, d *DABE, n,t int) *pbc.Element {
	id := d.CurveParam.GetNewZn().SetBytes(idb)
	sij := d.CurveParam.GetNewZn().Set0().ThenAdd(u.F[0])
	for index := 1; index < t; index++ {
		temp := d.CurveParam.GetNewZn().Set1().ThenMul(u.F[index])
		for i := 1; i <= index; i++ {
			temp.ThenMul(id)
		}
		sij.ThenAdd(temp)
	}
	return sij
}