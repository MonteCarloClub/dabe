package DecentralizedABE

import (
	"fmt"
	"github.com/Nik-U/pbc"
)

type Org struct {
	APKMap    map[string]*APK
	EGGAlpha  *pbc.Element
	Name      string
	N         int //总用户个数
	T         int //门限阈值
	UserNames map[string]*pbc.Element //用户的名称
}

func (o *Org) GetPK() *pbc.Element {
	return o.EGGAlpha
}

func (o *Org) GetAPKMap() map[string]*APK {
	return o.APKMap
}

//生成组织公钥
func (o *Org) GenerateOPK(names []string, pks []*pbc.Element, d *DABE) error {
	if len(pks) != o.T || len(names) != o.T{
		return fmt.Errorf("pks or names isn't eq t")
	}

	eGGAlpha := d.CurveParam.Get0FromGT().Set1()
	for i := 0 ; i< o.T ;i++ {
		up := d.CurveParam.Get0FromZn().Set1()
		for j := 0; j < o.T; j++ {
			if i == j {
				continue
			}
			di := d.CurveParam.Get0FromZn().Sub(o.UserNames[names[j]], o.UserNames[names[i]])
			di = d.CurveParam.GetNewZn().Div(o.UserNames[names[j]], di)
			up.ThenMul(di)
		}
		eGGAlpha.ThenMul(d.CurveParam.Get0FromGT().PowZn(pks[i], up))
	}
	o.EGGAlpha = eGGAlpha
	return nil
}

//生成属性
func (o *Org) GenerateNewAttr(names []string, apks []*pbc.Element, attr string, d *DABE) error {
	if len(apks) != o.T || len(names) != o.T{
		return fmt.Errorf("pks or names isn't eq t")
	}
	if o.APKMap[attr] != nil {
		return fmt.Errorf("already has this attr")
	}

	gY := d.CurveParam.Get0FromGT().Set1()
	for i := 0 ; i< o.T ;i++ {
		up := d.CurveParam.Get0FromZn().Set1()
		for j := 0; j < o.T; j++ {
			if i == j {
				continue
			}
			di := d.CurveParam.Get0FromZn().Sub(o.UserNames[names[j]], o.UserNames[names[i]])
			di = d.CurveParam.GetNewZn().Div(o.UserNames[names[j]], di)
			up.ThenMul(di)
		}
		gY.ThenMul(d.CurveParam.Get0FromGT().PowZn(apks[i], up))
	}
	o.APKMap[attr] = &APK{
		Gy: gY,
	}
	return nil
}