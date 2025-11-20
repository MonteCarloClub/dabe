package storage

import (
	"encoding/json"
	"fmt"

	model "github.com/MonteCarloClub/dabe/model"
	"github.com/Nik-U/pbc"
)

// DABESerialized DABE序列化结构
type DABESerialized struct {
	CurveParam *CurveParamSerialized `json:"curve_param"`
	G          string                `json:"g"`
	EGG        string                `json:"egg"`
}

// CurveParamSerialized 曲线参数序列化结构
type CurveParamSerialized struct {
	P      string `json:"p"`      // 大整数序列化为字符串
	Params []byte `json:"params"` // PBC参数序列化
}

// 序列化DABE系统参数
func SerializeDABE(dabe *model.DABE) ([]byte, error) {
	if dabe == nil {
		return nil, fmt.Errorf("DABE parameter is nil")
	}

	// 序列化G元素
	gString := dabe.G.String()

	// 序列化EGG元素
	eggString := dabe.EGG.String()

	// 构建序列化结构
	dabeData := DABESerialized{
		G:   gString,
		EGG: eggString,
	}

	// JSON序列化
	return json.Marshal(dabeData)
}

// 反序列化DABE系统参数
func DeserializeDABE(data []byte) (*model.DABE, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var dabeData DABESerialized
	if err := json.Unmarshal(data, &dabeData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DABE data: %v", err)
	}

	// 创建DABE实例
	dabe := &model.DABE{
		CurveParam: new(model.CurveParam),
	}
	dabe.CurveParam.Initialize()

	serializer := NewElementSerializer()
	// 反序列化G元素
	desG, err := serializer.DeserializeElement(dabeData.G, dabe.CurveParam, "G1")
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize G element: %v", err)
	}
	dabe.G = desG

	// 反序列化EGG元素
	desEGG, err := serializer.DeserializeElement(dabeData.EGG, dabe.CurveParam, "GT")
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize EGG element: %v", err)
	}
	dabe.EGG = desEGG
	return dabe, nil
}

// UserSerialized 用户序列化结构
type UserSerialized struct {
	Name     string                    `json:"name"`
	EGGAlpha []byte                    `json:"egg_alpha"`
	Alpha    []byte                    `json:"alpha"`
	GAlpha   []byte                    `json:"g_alpha"`
	APKMap   map[string]*APKSerialized `json:"apk_map"`
	ASKMap   map[string]*ASKSerialized `json:"ask_map"`
	OPKMap   map[string]*OPKSerialized `json:"opk_map,omitempty"`
	OSKMap   map[string]*OSKSerialized `json:"osk_map,omitempty"`
}

// APKSerialized 属性公钥序列化结构
type APKSerialized struct {
	Gy []byte `json:"gy"`
}

// ASKSerialized 属性私钥序列化结构
type ASKSerialized struct {
	Y []byte `json:"y"`
}

// OPKSerialized 组织公钥序列化结构
type OPKSerialized struct {
	OPK    []byte            `json:"opk"`
	APKMap map[string][]byte `json:"apk_map"`
}

// OSKSerialized 组织私钥序列化结构
type OSKSerialized struct {
	AlphaPart   []byte            `json:"alpha_part"`
	ASKMap      map[string][]byte `json:"ask_map"`
	F           [][]byte          `json:"f"`
	N           int               `json:"n"`
	T           int               `json:"t"`
	OthersShare [][]byte          `json:"others_share"`
	OSK         []byte            `json:"osk"`
	GOSK        []byte            `json:"gosk"`
}

// 序列化用户（权威机构）
func SerializeUser(user *model.User) ([]byte, error) {
	if user == nil {
		return nil, fmt.Errorf("user is nil")
	}

	userData := UserSerialized{
		Name:     user.Name,
		EGGAlpha: user.EGGAlpha.Bytes(),
		Alpha:    user.Alpha.Bytes(),
		GAlpha:   user.GAlpha.Bytes(),
		APKMap:   make(map[string]*APKSerialized),
		ASKMap:   make(map[string]*ASKSerialized),
	}

	// 序列化APK映射
	for attrName, apk := range user.APKMap {
		userData.APKMap[attrName] = &APKSerialized{
			Gy: apk.Gy.Bytes(),
		}
	}

	// 序列化ASK映射
	for attrName, ask := range user.ASKMap {
		userData.ASKMap[attrName] = &ASKSerialized{
			Y: ask.Y.Bytes(),
		}
	}

	// 序列化OPK映射（如果存在）
	if user.OPKMap != nil {
		userData.OPKMap = make(map[string]*OPKSerialized)
		for orgName, opk := range user.OPKMap {
			apkMap := make(map[string][]byte)
			for attrName, apkElem := range opk.APKMap {
				apkMap[attrName] = apkElem.Bytes()
			}
			userData.OPKMap[orgName] = &OPKSerialized{
				OPK:    opk.OPK.Bytes(),
				APKMap: apkMap,
			}
		}
	}

	// 序列化OSK映射（如果存在）
	if user.OSKMap != nil {
		userData.OSKMap = make(map[string]*OSKSerialized)
		for orgName, osk := range user.OSKMap {
			askMap := make(map[string][]byte)
			for attrName, askPart := range osk.ASKMap {
				askMap[attrName] = askPart.YPart.Bytes()
			}

			f := make([][]byte, len(osk.F))
			for i, fElem := range osk.F {
				f[i] = fElem.Bytes()
			}

			othersShare := make([][]byte, len(osk.OthersShare))
			for i, shareElem := range osk.OthersShare {
				othersShare[i] = shareElem.Bytes()
			}

			userData.OSKMap[orgName] = &OSKSerialized{
				AlphaPart:   osk.AlphaPart.Bytes(),
				ASKMap:      askMap,
				F:           f,
				N:           osk.N,
				T:           osk.T,
				OthersShare: othersShare,
				OSK:         osk.OSK.Bytes(),
				GOSK:        osk.GOSK.Bytes(),
			}
		}
	}

	return json.Marshal(userData)
}

// 反序列化用户（权威机构）
func DeserializeUser(data []byte, curveParam *model.CurveParam) (*model.User, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var userData UserSerialized
	if err := json.Unmarshal(data, &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %v", err)
	}

	// 创建用户实例
	user := &model.User{
		Name:   userData.Name,
		APKMap: make(map[string]*model.APK),
		ASKMap: make(map[string]*model.ASK),
	}

	// 反序列化基本元素
	user.EGGAlpha = curveParam.Get0FromGT()
	user.EGGAlpha.SetBytes(userData.EGGAlpha)

	user.Alpha = curveParam.Get0FromZn()
	user.Alpha.SetBytes(userData.Alpha)

	user.GAlpha = curveParam.Get0FromG1()
	user.GAlpha.SetBytes(userData.GAlpha)

	// 反序列化APK映射
	for attrName, apkData := range userData.APKMap {
		apk := &model.APK{}
		apk.Gy = curveParam.Get0FromG1()
		apk.Gy.SetBytes(apkData.Gy)
		user.APKMap[attrName] = apk
	}

	// 反序列化ASK映射
	for attrName, askData := range userData.ASKMap {
		ask := &model.ASK{}
		ask.Y = curveParam.Get0FromZn()
		ask.Y.SetBytes(askData.Y)
		user.ASKMap[attrName] = ask
	}

	// 反序列化OPK映射（如果存在）
	if userData.OPKMap != nil {
		user.OPKMap = make(map[string]*model.OPKPart)
		for orgName, opkData := range userData.OPKMap {
			opk := &model.OPKPart{
				APKMap: make(map[string]*pbc.Element),
			}

			opk.OPK = curveParam.Get0FromGT()
			opk.OPK.SetBytes(opkData.OPK)

			for attrName, apkBytes := range opkData.APKMap {
				apkElem := curveParam.Get0FromG1()
				apkElem.SetBytes(apkBytes)
				opk.APKMap[attrName] = apkElem
			}

			user.OPKMap[orgName] = opk
		}
	}

	// 反序列化OSK映射（如果存在）
	if userData.OSKMap != nil {
		user.OSKMap = make(map[string]*model.OSKPart)
		for orgName, oskData := range userData.OSKMap {
			osk := &model.OSKPart{
				ASKMap: make(map[string]*model.ASKPart),
				N:      oskData.N,
				T:      oskData.T,
			}

			osk.AlphaPart = curveParam.Get0FromZn()
			osk.AlphaPart.SetBytes(oskData.AlphaPart)

			// 反序列化ASK映射
			for attrName, askBytes := range oskData.ASKMap {
				askPart := &model.ASKPart{}
				askPart.YPart = curveParam.Get0FromZn()
				askPart.YPart.SetBytes(askBytes)
				osk.ASKMap[attrName] = askPart
			}

			// 反序列化F数组
			osk.F = make([]*pbc.Element, len(oskData.F))
			for i, fBytes := range oskData.F {
				osk.F[i] = curveParam.Get0FromZn()
				osk.F[i].SetBytes(fBytes)
			}

			// 反序列化OthersShare数组
			osk.OthersShare = make([]*pbc.Element, len(oskData.OthersShare))
			for i, shareBytes := range oskData.OthersShare {
				osk.OthersShare[i] = curveParam.Get0FromZn()
				osk.OthersShare[i].SetBytes(shareBytes)
			}

			osk.OSK = curveParam.Get0FromZn()
			osk.OSK.SetBytes(oskData.OSK)

			osk.GOSK = curveParam.Get0FromG1()
			osk.GOSK.SetBytes(oskData.GOSK)

			user.OSKMap[orgName] = osk
		}
	}

	return user, nil
}

// CipherSerialized 密文序列化结构
type CipherSerialized struct {
	C0         []byte   `json:"c0"`
	C1s        [][]byte `json:"c1s"`
	C2s        [][]byte `json:"c2s"`
	C3s        [][]byte `json:"c3s"`
	CipherText []byte   `json:"cipher_text"`
	Policy     string   `json:"policy"`
}

// 序列化密文
func SerializeCipher(cipher *model.Cipher) ([]byte, error) {
	if cipher == nil {
		return nil, fmt.Errorf("cipher is nil")
	}

	cipherData := CipherSerialized{
		C0:         cipher.C0.Bytes(),
		CipherText: cipher.CipherText,
		Policy:     cipher.Policy,
	}

	// 序列化C1s数组
	cipherData.C1s = make([][]byte, len(cipher.C1s))
	for i, c1 := range cipher.C1s {
		cipherData.C1s[i] = c1.Bytes()
	}

	// 序列化C2s数组
	cipherData.C2s = make([][]byte, len(cipher.C2s))
	for i, c2 := range cipher.C2s {
		cipherData.C2s[i] = c2.Bytes()
	}

	// 序列化C3s数组
	cipherData.C3s = make([][]byte, len(cipher.C3s))
	for i, c3 := range cipher.C3s {
		cipherData.C3s[i] = c3.Bytes()
	}

	return json.Marshal(cipherData)
}

// 反序列化密文
func DeserializeCipher(data []byte, curveParam *model.CurveParam) (*model.Cipher, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	var cipherData CipherSerialized
	if err := json.Unmarshal(data, &cipherData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cipher data: %v", err)
	}

	cipher := &model.Cipher{
		CipherText: cipherData.CipherText,
		Policy:     cipherData.Policy,
	}

	// 反序列化C0
	cipher.C0 = curveParam.Get0FromGT()
	cipher.C0.SetBytes(cipherData.C0)

	// 反序列化C1s数组
	cipher.C1s = make([]*pbc.Element, len(cipherData.C1s))
	for i, c1Bytes := range cipherData.C1s {
		cipher.C1s[i] = curveParam.Get0FromGT()
		cipher.C1s[i].SetBytes(c1Bytes)
	}

	// 反序列化C2s数组
	cipher.C2s = make([]*pbc.Element, len(cipherData.C2s))
	for i, c2Bytes := range cipherData.C2s {
		cipher.C2s[i] = curveParam.Get0FromG1()
		cipher.C2s[i].SetBytes(c2Bytes)
	}

	// 反序列化C3s数组
	cipher.C3s = make([]*pbc.Element, len(cipherData.C3s))
	for i, c3Bytes := range cipherData.C3s {
		cipher.C3s[i] = curveParam.Get0FromG1()
		cipher.C3s[i].SetBytes(c3Bytes)
	}

	return cipher, nil
}
