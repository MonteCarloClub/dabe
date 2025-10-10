package test

import (
	"fmt"
	"testing"

	"github.com/MonteCarloClub/dabe/model"
	"github.com/MonteCarloClub/dabe/storage"
	"github.com/Nik-U/pbc"
)

func TestStorageBasic(t *testing.T) {
	//初始化和全局参数生成
	dabe := new(model.DABE)
	dabe.GlobalSetup()

	// 持久化全局参数
	fs, err := storage.NewFileStorage("storage")
	if err != nil {
		panic(err)
	}
	err = fs.SaveGlobalParams(dabe)
	if err != nil {
		panic(err)
	}

	// 模拟重启恢复全局参数
	loadedDABE, err := fs.LoadGlobalParams()
	if err != nil {
		panic(err)
	}
	dabe = loadedDABE

	//初始化两个不同的权限管理机构，并保存
	authorityPKMap := make(map[string]*pbc.Element)
	authorityMap := make(map[string]model.Authority)
	fudanUniversity := dabe.UserSetup("Fudan_University")
	authorityPKMap["Fudan_University"] = fudanUniversity.GetPK()
	// authorityMap["Fudan_University"] = fudanUniversity
	ageAuthority := dabe.UserSetup("Age_Authority")
	authorityPKMap["Age_Authority"] = ageAuthority.GetPK()
	// authorityMap["Age_Authority"] = ageAuthority

	// 保存机构
	err = fs.SaveUserAuthority(fudanUniversity, "Fudan_University")
	if err != nil {
		panic(err)
	}
	err = fs.SaveUserAuthority(ageAuthority, "Age_Authority")
	if err != nil {
		panic(err)
	}

	// 重新加载机构，验证持久化正确性
	fudanUniversity, err = fs.LoadUserAuthority("Fudan_University", dabe.CurveParam)
	if err != nil {
		panic(err)
	}
	ageAuthority, err = fs.LoadUserAuthority("Age_Authority", dabe.CurveParam)
	if err != nil {
		panic(err)
	}
	authorityMap["Fudan_University"] = fudanUniversity
	authorityMap["Age_Authority"] = ageAuthority

	//保存所有属性公钥
	pkMap := make(map[string]*model.APK)
	//生成属性公钥
	tempPk, err := fudanUniversity.GenerateNewAttr("Fudan_University:在读研究生", dabe)
	if err != nil {
		panic(err)
	}
	pkMap["Fudan_University:在读研究生"] = tempPk
	tempPk2, err := ageAuthority.GenerateNewAttr("Age_Authority:23", dabe)
	if err != nil {
		panic(err)
	}
	pkMap["Age_Authority:23"] = tempPk2
	tempPk3, err := ageAuthority.GenerateNewAttr("Age_Authority:24", dabe)
	if err != nil {
		panic(err)
	}
	pkMap["Age_Authority:24"] = tempPk3
	// 持久化属性公钥
	err = fs.SaveAttributePublicKeys(pkMap)
	if err != nil {
		panic(err)
	}
	// 重新加载属性公钥，验证持久化正确性
	loadedPkMap, err := fs.LoadAttributePublicKeys(dabe.CurveParam)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded %d attribute public keys from storage\n", len(loadedPkMap))

	//用户申请密钥
	user1Privatekeys := make(map[string]*pbc.Element)
	user2Privatekeys := make(map[string]*pbc.Element)

	user1Privatekey1, err := fudanUniversity.KeyGenByUser("陈泽宁", "Fudan_University:在读研究生", dabe)
	if err != nil {
		panic(err)
	}
	user1Privatekey2, err := ageAuthority.KeyGenByUser("陈泽宁", "Age_Authority:23", dabe)
	if err != nil {
		panic(err)
	}
	user2Privatekey1, err := ageAuthority.KeyGenByUser("24岁的无名氏", "Age_Authority:24", dabe)
	if err != nil {
		panic(err)
	}
	user1Privatekeys["Fudan_University:在读研究生"] = user1Privatekey1
	user1Privatekeys["Age_Authority:23"] = user1Privatekey2
	user2Privatekeys["Age_Authority:24"] = user2Privatekey1

	//加密两个不同的明文,这里authorityMap应该不传入私钥相关，方便起见如此做
	m1 := "复旦的在读研究生或者24岁的人可以看见"
	m2 := "复旦的23岁在读研究生可以看见"
	cipher1, err := dabe.EncryptWithKeys(m1, "(Fudan_University:在读研究生 OR Age_Authority:24)", pkMap, authorityPKMap)
	if err != nil {
		panic(err)
	}
	cipher2, err := dabe.EncryptWithKeys(m2, "(Fudan_University:在读研究生 AND Age_Authority:23)", pkMap, authorityPKMap)
	if err != nil {
		panic(err)
	}

	//解密
	decrypt, err := dabe.Decrypt(cipher1, user1Privatekeys, "陈泽宁")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("陈泽宁解密出了： " + string(decrypt))
	}
	decrypt2, err := dabe.Decrypt(cipher2, user1Privatekeys, "陈泽宁")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("陈泽宁解密出了： " + string(decrypt2))
	}
	decrypt3, err := dabe.Decrypt(cipher1, user2Privatekeys, "24岁的无名氏")
	if err != nil {
		panic(err)
	} else {
		fmt.Println("24岁的无名氏解密出了： " + string(decrypt3))
	}
	decrypt4, err := dabe.Decrypt(cipher2, user2Privatekeys, "24岁的无名氏")
	if err == nil {
		fmt.Println("24岁的无名氏错误解密出了： " + string(decrypt4))
	} else {
		fmt.Println("24岁的无名氏正常地失败于： " + err.Error())
	}
}

// TestSerialization 测试DABE序列化功能
func TestSerialization(t *testing.T) {
	t.Log("开始DABE序列化测试...")

	// 创建临时存储目录
	fs, err := storage.NewFileStorage("test_serialization")
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 1. 测试DABE全局参数序列化
	t.Log("测试DABE全局参数序列化...")
	dabe := &model.DABE{}
	dabe.GlobalSetup()

	// 保存全局参数
	if err := fs.SaveGlobalParams(dabe); err != nil {
		t.Fatalf("保存全局参数失败: %v", err)
	}
	t.Log("✓ 全局参数保存成功")

	// 加载全局参数
	loadedDABE, err := fs.LoadGlobalParams()
	if err != nil {
		t.Fatalf("加载全局参数失败: %v", err)
	}
	t.Log("✓ 全局参数加载成功")

	// 验证加载的参数
	if loadedDABE.G.Bytes() == nil || loadedDABE.EGG.Bytes() == nil {
		t.Fatal("加载的全局参数无效")
	}
	if len(loadedDABE.G.Bytes()) != len(dabe.G.Bytes()) {
		t.Fatal("G元素序列化不一致")
	}
	if len(loadedDABE.EGG.Bytes()) != len(dabe.EGG.Bytes()) {
		t.Fatal("EGG元素序列化不一致")
	}
	t.Log("✓ 全局参数验证成功")

	// 2. 测试用户权威机构序列化
	t.Log("测试用户权威机构序列化...")

	// 创建用户权威机构
	authority := dabe.UserSetup("TestAuthority")

	// 为权威机构添加属性
	_, err = authority.GenerateNewAttr("TestAuthority:Student", dabe)
	if err != nil {
		t.Fatalf("生成属性失败: %v", err)
	}

	_, err = authority.GenerateNewAttr("TestAuthority:Teacher", dabe)
	if err != nil {
		t.Fatalf("生成属性失败: %v", err)
	}

	originalAttrCount := len(authority.APKMap)

	// 保存用户权威机构
	if err := fs.SaveUserAuthority(authority, "TestAuthority"); err != nil {
		t.Fatalf("保存用户权威机构失败: %v", err)
	}
	t.Log("✓ 用户权威机构保存成功")

	// 加载用户权威机构
	loadedAuthority, err := fs.LoadUserAuthority("TestAuthority", dabe.CurveParam)
	if err != nil {
		t.Fatalf("加载用户权威机构失败: %v", err)
	}
	t.Log("✓ 用户权威机构加载成功")

	// 验证加载的权威机构
	if loadedAuthority.Name != authority.Name {
		t.Fatalf("权威机构名称不匹配: 原始=%s, 加载=%s",
			authority.Name, loadedAuthority.Name)
	}
	if len(loadedAuthority.APKMap) != originalAttrCount {
		t.Fatalf("权威机构属性数量不匹配: 原始=%d, 加载=%d",
			originalAttrCount, len(loadedAuthority.APKMap))
	}

	// 验证具体属性
	for attrName := range authority.APKMap {
		if _, exists := loadedAuthority.APKMap[attrName]; !exists {
			t.Fatalf("属性 %s 在加载的权威机构中不存在", attrName)
		}
	}
	t.Logf("✓ 权威机构验证成功，属性数量: %d", len(loadedAuthority.APKMap))

	// 3. 测试序列化接口方法
	t.Log("测试序列化接口方法...")

	// 测试全局参数序列化接口
	globalParamsJSON, err := fs.SerializeGlobalParams(dabe)
	if err != nil {
		t.Fatalf("序列化全局参数失败: %v", err)
	}

	if len(globalParamsJSON) == 0 {
		t.Fatal("序列化的全局参数为空")
	}
	t.Logf("全局参数JSON大小: %d bytes", len(globalParamsJSON))

	// 从JSON加载全局参数
	loadedDABE2, err := fs.LoadGlobalParamsFromJSON(globalParamsJSON)
	if err != nil {
		t.Fatalf("从JSON加载全局参数失败: %v", err)
	}
	t.Log("✓ 全局参数JSON序列化测试成功")

	// 测试用户权威机构序列化接口
	userJSON, err := fs.SerializeUserAuthority(authority)
	if err != nil {
		t.Fatalf("序列化用户权威机构失败: %v", err)
	}

	if len(userJSON) == 0 {
		t.Fatal("序列化的用户权威机构为空")
	}
	t.Logf("用户权威机构JSON大小: %d bytes", len(userJSON))

	loadedUser, err := fs.DeserializeUserAuthority(userJSON, dabe.CurveParam)
	if err != nil {
		t.Fatalf("反序列化用户权威机构失败: %v", err)
	}

	if loadedUser.Name != authority.Name {
		t.Fatal("用户权威机构序列化验证失败")
	}
	t.Log("✓ 用户权威机构JSON序列化测试成功")

	// 4. 验证系统兼容性
	t.Log("验证系统兼容性...")

	// 比较曲线参数
	if loadedDABE.CurveParam.GetP().Cmp(loadedDABE2.CurveParam.GetP()) != 0 {
		t.Fatal("两个反序列化DABE系统的曲线参数不一致")
	}
	t.Log("✓ 曲线参数一致性验证成功")

	// 比较G和EGG元素
	if !loadedDABE.G.Equals(loadedDABE2.G) {
		t.Fatal("两个反序列化DABE系统的G元素不一致")
	}
	if !loadedDABE.EGG.Equals(loadedDABE2.EGG) {
		t.Fatal("两个反序列化DABE系统的EGG元素不一致")
	}
	t.Log("✓ G和EGG元素一致性验证成功")

	// 5. 功能性测试 - 使用反序列化的组件进行密钥生成
	t.Log("功能性测试...")

	// 使用反序列化的权威机构生成私钥
	privateKey, err := loadedUser.KeyGenByUser("TestUser", "TestAuthority:Student", loadedDABE2)
	if err != nil {
		t.Fatalf("使用反序列化组件生成私钥失败: %v", err)
	}

	if privateKey == nil {
		t.Fatal("生成的私钥为空")
	}
	if len(privateKey.Bytes()) == 0 {
		t.Fatal("生成的私钥数据为空")
	}

	t.Log("✓ 使用反序列化组件生成私钥成功")

	t.Log("=== 所有序列化测试通过 ===")
}

// TestElementSerialization 测试PBC元素序列化
func TestElementSerialization(t *testing.T) {
	t.Log("开始PBC元素序列化测试...")

	// 初始化DABE系统
	dabe := &model.DABE{}
	dabe.GlobalSetup()

	// 创建元素序列化器
	serializer := storage.NewElementSerializer()

	// 测试G1元素序列化
	t.Log("测试G1元素序列化...")
	g1Element := dabe.CurveParam.GetNewG1()
	g1Bytes, err := serializer.SerializeElement(g1Element)
	if err != nil {
		t.Fatalf("序列化G1元素失败: %v", err)
	}
	if len(g1Bytes) == 0 {
		t.Fatal("G1元素序列化结果为空")
	}

	// 反序列化G1元素
	deserializedG1, err := serializer.DeserializeElement(g1Bytes, dabe.CurveParam, "G1")
	if err != nil {
		t.Fatalf("反序列化G1元素失败: %v", err)
	}

	if !g1Element.Equals(deserializedG1) {
		t.Fatal("G1元素序列化后不一致")
	}
	t.Log("✓ G1元素序列化测试成功")

	// 测试GT元素序列化
	t.Log("测试GT元素序列化...")
	gtElement := dabe.CurveParam.GetNewGT()
	gtBytes, err := serializer.SerializeElement(gtElement)
	if err != nil {
		t.Fatalf("序列化GT元素失败: %v", err)
	}

	deserializedGT, err := serializer.DeserializeElement(gtBytes, dabe.CurveParam, "GT")
	if err != nil {
		t.Fatalf("反序列化GT元素失败: %v", err)
	}

	if !gtElement.Equals(deserializedGT) {
		t.Fatal("GT元素序列化后不一致")
	}
	t.Log("✓ GT元素序列化测试成功")

	// 测试Zr元素序列化
	t.Log("测试Zr元素序列化...")
	zrElement := dabe.CurveParam.GetNewZn()
	zrBytes, err := serializer.SerializeElement(zrElement)
	if err != nil {
		t.Fatalf("序列化Zr元素失败: %v", err)
	}

	deserializedZr, err := serializer.DeserializeElement(zrBytes, dabe.CurveParam, "Zr")
	if err != nil {
		t.Fatalf("反序列化Zr元素失败: %v", err)
	}

	if !zrElement.Equals(deserializedZr) {
		t.Fatal("Zr元素序列化后不一致")
	}
	t.Log("✓ Zr元素序列化测试成功")

	// 测试空元素处理
	t.Log("测试空元素处理...")
	nilBytes, err := serializer.SerializeElement(nil)
	if err != nil {
		t.Fatalf("序列化空元素失败: %v", err)
	}
	if nilBytes != nil {
		t.Fatal("空元素序列化应该返回nil")
	}

	nilElement, err := serializer.DeserializeElement(nil, dabe.CurveParam, "G1")
	if err != nil {
		t.Fatalf("反序列化空数据失败: %v", err)
	}
	if nilElement != nil {
		t.Fatal("空数据反序列化应该返回nil")
	}
	t.Log("✓ 空元素处理测试成功")

	t.Log("=== PBC元素序列化测试通过 ===")
}

// TestFileStorageMethods 测试文件存储的各种方法
func TestFileStorageMethods(t *testing.T) {
	t.Log("开始文件存储方法测试...")

	// 创建临时存储
	fs, err := storage.NewFileStorage("test_file_methods")
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	// 初始化DABE系统
	dabe := &model.DABE{}
	dabe.GlobalSetup()

	// 测试多次保存和加载
	t.Log("测试多次保存和加载...")
	for i := 0; i < 3; i++ {
		// 保存
		if err := fs.SaveGlobalParams(dabe); err != nil {
			t.Fatalf("第%d次保存全局参数失败: %v", i+1, err)
		}

		// 加载
		loadedDABE, err := fs.LoadGlobalParams()
		if err != nil {
			t.Fatalf("第%d次加载全局参数失败: %v", i+1, err)
		}

		// 验证
		if !dabe.G.Equals(loadedDABE.G) || !dabe.EGG.Equals(loadedDABE.EGG) {
			t.Fatalf("第%d次验证失败", i+1)
		}
	}
	t.Log("✓ 多次保存和加载测试成功")

	// 测试权威机构的多个实例
	t.Log("测试多个权威机构...")
	authorities := make([]*model.User, 3)
	authorityNames := []string{"University", "Government", "Company"}

	// 创建多个权威机构
	for i, name := range authorityNames {
		authorities[i] = dabe.UserSetup(name)

		// 为每个权威机构添加属性
		_, err = authorities[i].GenerateNewAttr(fmt.Sprintf("%s:Member", name), dabe)
		if err != nil {
			t.Fatalf("为%s生成属性失败: %v", name, err)
		}

		// 保存权威机构
		if err := fs.SaveUserAuthority(authorities[i], name); err != nil {
			t.Fatalf("保存权威机构%s失败: %v", name, err)
		}
	}

	// 加载并验证所有权威机构
	for i, name := range authorityNames {
		loadedAuth, err := fs.LoadUserAuthority(name, dabe.CurveParam)
		if err != nil {
			t.Fatalf("加载权威机构%s失败: %v", name, err)
		}

		if loadedAuth.Name != authorities[i].Name {
			t.Fatalf("权威机构%s名称不匹配", name)
		}

		if len(loadedAuth.APKMap) != len(authorities[i].APKMap) {
			t.Fatalf("权威机构%s属性数量不匹配", name)
		}
	}
	t.Log("✓ 多个权威机构测试成功")

	t.Log("=== 文件存储方法测试通过 ===")
}
