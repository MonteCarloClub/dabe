package DecentralizedABE

import (
	"crypto/sha256"
	"fmt"
	"github.com/Nik-U/pbc"
)

type User struct {
	APKMap   map[string]*APK
	ASKMap   map[string]*ASK
	EGGAlpha *pbc.Element
	Alpha    *pbc.Element
	Name     string
	OPKMap   map[string]*OPKPart
	OSKMap   map[string]*OSKPart
	KeyMap   map[string]*pbc.Element
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

//授权用户属性
func (u *User) KeyGenByUser(gid string, attr string, d *DABE) (*pbc.Element, error) {
	if u.ASKMap[attr] == nil {
		return nil, fmt.Errorf("don't have this attr, error when %s", attr)
	}
	temp := sha256.Sum256([]byte(gid))
	hashGid := d.G.NewFieldElement().SetBytes(temp[:])
	key := d.G.NewFieldElement().PowZn(d.G, u.Alpha).ThenMul(hashGid.ThenPowZn(u.ASKMap[attr].Y))
	return key, nil
}

//授权组织属性
func (u *User) KeyGenByOrg(gid string, attr string, d *DABE, orgName string) (*pbc.Element, error) {
	if u.OSKMap[orgName] == nil || u.OSKMap[orgName].ASKPartMap[attr] == nil {
		return nil, fmt.Errorf("don't have this attr, error when %s", attr)
	}
	temp := sha256.Sum256([]byte(gid))
	hashGid := d.G.NewFieldElement().SetBytes(temp[:])
	key := d.G.NewFieldElement().Set(u.OSKMap[orgName].GAlphaPart).
		ThenMul(hashGid.ThenPowZn(u.OSKMap[orgName].ASKPartMap[attr].YPart))
	return key, nil
}

//创建Org所需的秘密share
func (u *User) GenerateOrgShare(n, t int, userNames map[string]*pbc.Element, orgName string, d *DABE) (
	map[string]*pbc.Element, error) {

	if u.OSKMap[orgName] != nil || u.OPKMap[orgName] != nil {
		return nil, fmt.Errorf("already has this org")
	}
	alphaPart := d.CurveParam.GetNewZn()
	gAlphaPart := d.CurveParam.Get0FromG1().PowZn(d.G, alphaPart)
	f := make([]*pbc.Element, 0, 0)
	f = append(f, alphaPart)
	for i := 1; i < t; i++ {
		f = append(f, d.CurveParam.GetNewZn())
	}
	oskPart := &OSKPart{
		AlphaPart:  alphaPart,
		F:          f,
		N:          n,
		T:          t,
		GAlphaPart: gAlphaPart,
	}
	opkPart := &OPKPart{
		EGGAlphaPart: d.CurveParam.Get0FromGT().PowZn(d.EGG, alphaPart),
	}
	u.OSKMap[orgName] = oskPart
	u.OPKMap[orgName] = opkPart

	shares := make(map[string]*pbc.Element)
	for name, hGID := range userNames {
		shares[name] = u.share(hGID, d, n, t, f)
	}
	return shares, nil
}

//创建Org属性所需的秘密share
func (u *User) GenerateOrgAttrShare(n, t int, org *Org, d *DABE, attrName string) (
	map[string]*pbc.Element, error) {

	if !CheckAttrName(attrName, org.Name) {
		return nil, fmt.Errorf("attrName is invalid")
	}
	if u.OSKMap[org.Name] == nil || u.OPKMap[org.Name] == nil {
		return nil, fmt.Errorf("doesn't has this org")
	}
	if u.OSKMap[org.Name].ASKPartMap[attrName] != nil || u.OPKMap[org.Name].GyPart[attrName] != nil {
		return nil, fmt.Errorf("already has this attr")
	}
	yPart := d.CurveParam.GetNewZn()
	f := make([]*pbc.Element, 0, 0)
	f = append(f, yPart)
	for i := 1; i < t; i++ {
		f = append(f, d.CurveParam.GetNewZn())
	}

	askPart := &ASKPart{
		F:     f,
		YPart: yPart,
	}
	u.OPKMap[org.Name].GyPart[attrName] = d.CurveParam.Get0FromG1().PowZn(d.G, yPart)
	u.OSKMap[org.Name].ASKPartMap[attrName] = askPart

	shares := make(map[string]*pbc.Element)
	for name, hGID := range org.UserNames {
		shares[name] = u.share(hGID, d, n, t, f)
	}
	return shares, nil
}

//组装组织属性
func (u *User) AssembleOrgAttr(names []string, name2attr map[string]*pbc.Element,
	name2gid map[string]*pbc.Element, d *DABE, t int, attr string) error {

	if len(name2attr) != t || len(name2gid) != t {
		return fmt.Errorf("length not enough")
	}
	key := d.CurveParam.Get1FromG1()

	for i := 0; i < t; i++ {
		up := d.CurveParam.Get1FromZn()
		for j := 0; j < t; j++ {
			if i == j {
				continue
			}
			di := d.CurveParam.Get0FromZn().Sub(name2gid[names[j]], name2gid[names[i]])
			di = d.CurveParam.GetNewZn().Div(name2gid[names[j]], di)
			up.ThenMul(di)
		}
		key.ThenMul(d.CurveParam.Get0FromGT().PowZn(name2attr[names[i]], up))
	}
	u.KeyMap[attr] = key
	return nil
}

//get sij
func (u *User) share(otherHGID *pbc.Element, d *DABE, n, t int, f []*pbc.Element) *pbc.Element {
	sij := d.CurveParam.Get0FromZn()
	//from t-1 -> 1, O(t)
	for index := t - 1; index >= 1; index-- {
		sij.ThenAdd(f[index])
		sij.ThenMul(otherHGID)
	}
	sij.ThenAdd(f[0])
	return sij
}
