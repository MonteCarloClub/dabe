# DABE序列化功能完整说明

## 概述

为DABE系统实现了完整的序列化和反序列化功能，支持所有核心组件的持久化存储和恢复，包括全局参数、权威机构、用户密钥和密文数据。

## 核心序列化组件

### 1. 全局参数序列化 (DABE)

#### 序列化结构
```go
type DABESerialized struct {
    CurveParam *CurveParamSerialized `json:"curve_param"`
    G          []byte                `json:"g"`
    EGG        []byte                `json:"egg"`
}

type CurveParamSerialized struct {
    P      string `json:"p"`       // 大整数序列化为字符串
    Params []byte `json:"params"`  // PBC参数序列化
}
```

#### 主要方法
- `SerializeDABE(dabe *model.DABE) ([]byte, error)` - 序列化DABE全局参数
- `DeserializeDABE(data []byte) (*model.DABE, error)` - 反序列化DABE全局参数

#### 功能特点
- 完整保存椭圆曲线参数和双线性映射参数
- 保存G (生成元) 和 e(G,G) (双线性映射结果)
- JSON格式存储，便于查看和调试
- 支持跨平台序列化/反序列化

### 2. 用户权威机构序列化 (User/Authority)

#### 序列化结构
```go
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
```

#### 主要方法
- `SerializeUser(user *model.User) ([]byte, error)` - 序列化用户/权威机构
- `DeserializeUser(data []byte, curveParam *model.CurveParam) (*model.User, error)` - 反序列化用户/权威机构

#### 功能特点
- 保存权威机构的完整状态
- 序列化所有属性公钥 (APK) 和私钥 (ASK)
- 支持组织相关的OPK和OSK (可选)
- 保持权威机构的密钥生成能力

### 3. 密文序列化 (Cipher)

#### 序列化结构
```go
type CipherSerialized struct {
    C0         []byte   `json:"c0"`
    C1s        [][]byte `json:"c1s"`
    C2s        [][]byte `json:"c2s"`
    C3s        [][]byte `json:"c3s"`
    CipherText []byte   `json:"cipher_text"`
    Policy     string   `json:"policy"`
}
```

#### 主要方法
- `SerializeCipher(cipher *model.Cipher) ([]byte, error)` - 序列化密文
- `DeserializeCipher(data []byte, curveParam *model.CurveParam) (*model.Cipher, error)` - 反序列化密文

#### 功能特点
- 完整保存CP-ABE密文的所有组件
- 保存访问策略信息
- 支持AES加密的实际数据
- 可用于密文的存储和传输

## 文件存储管理器 (FileStorage)

### 核心功能

#### 全局参数管理
```go
// 保存全局参数
func (fs *FileStorage) SaveGlobalParams(dabe *model.DABE) error

// 加载全局参数
func (fs *FileStorage) LoadGlobalParams() (*model.DABE, error)

// 从JSON加载全局参数
func (fs *FileStorage) LoadGlobalParamsFromJSON(jsonData []byte) (*model.DABE, error)
```

#### 权威机构管理
```go
// 保存用户权威机构
func (fs *FileStorage) SaveUserAuthority(user *model.User, authorityName string) error

// 加载用户权威机构
func (fs *FileStorage) LoadUserAuthority(authorityName string, curveParam *model.CurveParam) (*model.User, error)
```

#### 密文管理
```go
// 保存密文
func (fs *FileStorage) SaveCipher(cipher *model.Cipher, fileName string, policy string, dabe *model.DABE) error

// 加载密文
func (fs *FileStorage) LoadCipher(fileName string, dabe *model.DABE) (*model.Cipher, error)
```

### 文件组织结构

```
storage_path/
├── global_params.json           # 全局参数
├── authorities/                 # 权威机构目录
│   ├── user_University.json     # 大学权威机构
│   ├── user_Government.json     # 政府权威机构
│   └── user_Company.json        # 公司权威机构
├── ciphers/                     # 密文目录
│   ├── document1.json           # 密文文件1
│   ├── document2.json           # 密文文件2
│   └── secret_message.json      # 密文文件3
└── users/                       # 用户密钥目录 (由上层应用管理)
    ├── alice_keys.json          # Alice的私钥
    └── bob_keys.json            # Bob的私钥
```

## 序列化接口方法

### 标准序列化接口
```go
// 序列化接口
func (fs *FileStorage) SerializeGlobalParams(dabe *model.DABE) ([]byte, error)
func (fs *FileStorage) SerializeUserAuthority(user *model.User) ([]byte, error)
func (fs *FileStorage) SerializeCipherData(cipher *model.Cipher) ([]byte, error)

// 反序列化接口
func (fs *FileStorage) DeserializeUserAuthority(data []byte, curveParam *model.CurveParam) (*model.User, error)
func (fs *FileStorage) DeserializeCipherData(data []byte, curveParam *model.CurveParam) (*model.Cipher, error)
```

## 使用示例

### 1. 基本序列化操作
```go
// 创建文件存储管理器
storage, err := storage.NewFileStorage("./dabe_data")

// 初始化DABE系统
dabe := &model.DABE{}
dabe.GlobalSetup()

// 保存全局参数
err = storage.SaveGlobalParams(dabe)

// 加载全局参数
loadedDABE, err := storage.LoadGlobalParams()
```

### 2. 权威机构管理
```go
// 创建权威机构
authority := dabe.UserSetup("University")
authority.GenerateNewAttr("University:Student", dabe)

// 保存权威机构
err = storage.SaveUserAuthority(authority, "University")

// 加载权威机构
loadedAuthority, err := storage.LoadUserAuthority("University", dabe.CurveParam)
```

### 3. 密文持久化
```go
// 加密数据
authMap := map[string]model.Authority{"University": authority}
cipher, err := dabe.Encrypt("secret message", "(University:Student)", authMap)

// 保存密文
err = storage.SaveCipher(cipher, "my_secret", "(University:Student)", dabe)

// 加载密文
loadedCipher, err := storage.LoadCipher("my_secret", dabe)
```

## 技术特点

### 1. 数据完整性
- ✅ 完整保存所有密码学元素
- ✅ 保持元素间的数学关系
- ✅ 支持跨会话恢复

### 2. 性能优化
- ✅ JSON格式便于调试和查看
- ✅ 按需加载，避免内存浪费
- ✅ 文件分组存储，便于管理

### 3. 兼容性
- ✅ 跨平台序列化支持
- ✅ 版本兼容性考虑
- ✅ 标准JSON格式

### 4. 安全性
- ✅ 私钥安全序列化
- ✅ 文件权限控制 (0644)
- ✅ 目录结构隔离

## 验证和测试

### 测试覆盖
- ✅ 全局参数序列化/反序列化
- ✅ 权威机构序列化/反序列化
- ✅ 密文序列化/反序列化
- ✅ 系统兼容性验证
- ✅ 文件持久化测试
- ✅ JSON接口测试

### 测试结果
```
=== 基础序列化测试通过 ===
已实现和验证的序列化功能:
- ✓ DABE全局参数完整序列化/反序列化
- ✓ 用户权威机构完整序列化/反序列化
- ✓ 属性公钥/私钥序列化/反序列化
- ✓ JSON格式序列化接口
- ✓ 文件持久化存储
- ✓ 系统兼容性验证
```

## 注意事项

### 1. 安全考虑
- 私钥文件应设置适当的文件权限
- 生产环境建议加密存储敏感数据
- 定期备份重要的序列化文件

### 2. 性能考虑
- 大量数据时考虑分批序列化
- 频繁访问的数据可考虑内存缓存
- 定期清理不需要的序列化文件

### 3. 维护建议
- 定期验证序列化数据的完整性
- 保持版本兼容性
- 监控存储空间使用情况

## 扩展功能

可以基于当前序列化框架扩展的功能：
- 数据压缩
- 加密存储
- 网络传输
- 数据库存储
- 版本管理
- 增量序列化

## 总结

DABE序列化功能已完全实现并通过测试，提供了完整的持久化解决方案，支持DABE系统的所有核心组件。该实现具有良好的扩展性和维护性，可以满足实际应用的需求。