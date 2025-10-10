package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	model "github.com/MonteCarloClub/dabe/model"
	"github.com/Nik-U/pbc"
)

// ElementSerializer PBC元素序列化器
type ElementSerializer struct{}

// NewElementSerializer 创建新的元素序列化器
func NewElementSerializer() *ElementSerializer {
	return &ElementSerializer{}
}

// SerializeElement 序列化PBC元素为字节数组
func (es *ElementSerializer) SerializeElement(element *pbc.Element) ([]byte, error) {
	if element == nil {
		return nil, nil
	}
	return element.Bytes(), nil
}

// DeserializeElement 反序列化PBC元素
func (es *ElementSerializer) DeserializeElement(data []byte, curveParam *model.CurveParam, fieldType string) (*pbc.Element, error) {
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	var element *pbc.Element
	switch fieldType {
	case "G1":
		element = curveParam.Get0FromG1()
	case "GT":
		element = curveParam.Get0FromGT()
	case "Zr":
		element = curveParam.Get0FromZn()
	default:
		return nil, fmt.Errorf("unsupported field type: %s", fieldType)
	}

	element.SetBytes(data)
	return element, nil
}

// FileStorage 文件存储管理器
type FileStorage struct {
	basePath   string
	serializer *ElementSerializer
}

// NewFileStorage 创建新的文件存储管理器
func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	return &FileStorage{
		basePath:   basePath,
		serializer: NewElementSerializer(),
	}, nil
}

// AuthorityFile 权威机构文件结构
type AuthorityFile struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	PublicKey    []byte            `json:"public_key"`
	Attributes   map[string][]byte `json:"attributes"`
	RegisterTime string            `json:"register_time"`
}

// CipherFile 密文文件结构
type CipherFile struct {
	FileName   string `json:"file_name"`
	Policy     string `json:"policy"`
	CipherData []byte `json:"cipher_data"`
	CreateTime string `json:"create_time"`
}

// ============ 全局参数持久化 ============

// SaveGlobalParams 持久化全局参数
func (fs *FileStorage) SaveGlobalParams(dabe *model.DABE) error {
	if dabe == nil {
		return fmt.Errorf("DABE parameter is nil")
	}

	// 序列化DABE参数
	data, err := SerializeDABE(dabe)
	if err != nil {
		return fmt.Errorf("failed to serialize DABE: %v", err)
	}

	// 保存到文件
	globalParamPath := filepath.Join(fs.basePath, "global_params.json")
	if err := os.WriteFile(globalParamPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write global params file: %v", err)
	}

	return nil
}

// LoadGlobalParams 读取全局参数
func (fs *FileStorage) LoadGlobalParams() (*model.DABE, error) {
	globalParamPath := filepath.Join(fs.basePath, "global_params.json")

	// 检查文件是否存在
	if _, err := os.Stat(globalParamPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("global params file does not exist")
	}

	// 读取文件
	data, err := os.ReadFile(globalParamPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read global params file: %v", err)
	}

	// 反序列化
	dabe, err := DeserializeDABE(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize DABE: %v", err)
	}

	return dabe, nil
}

// LoadGlobalParamsFromJSON 从JSON数据加载全局参数
func (fs *FileStorage) LoadGlobalParamsFromJSON(jsonData []byte) (*model.DABE, error) {
	return DeserializeDABE(jsonData)
}

// ============ 权威机构持久化 ============

// SaveAuthority 保存权威机构
func (fs *FileStorage) SaveAuthority(authority model.Authority, authorityType string) error {
	var attributes map[string][]byte

	// 序列化属性公钥
	if authority.GetAPKMap() != nil {
		attributes = make(map[string][]byte)
		for attrName, apk := range authority.GetAPKMap() {
			apkBytes, err := fs.serializer.SerializeElement(apk.Gy)
			if err != nil {
				return err
			}
			attributes[attrName] = apkBytes
		}
	}

	// 序列化权威机构公钥
	pkBytes, err := fs.serializer.SerializeElement(authority.GetPK())
	if err != nil {
		return err
	}

	var authorityName string
	switch auth := authority.(type) {
	case *model.User:
		authorityName = auth.Name
	case *model.Org:
		authorityName = auth.Name
	default:
		return fmt.Errorf("unsupported authority type")
	}

	authorityFile := AuthorityFile{
		Name:         authorityName,
		Type:         authorityType,
		PublicKey:    pkBytes,
		Attributes:   attributes,
		RegisterTime: getCurrentTimestamp(),
	}

	return fs.saveToFile(fmt.Sprintf("authorities/%s_%s.json", authorityType, authorityName), authorityFile)
}

// SaveUserAuthority 保存用户权威机构的完整信息
func (fs *FileStorage) SaveUserAuthority(user *model.User, authorityName string) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}

	// 序列化用户
	userData, err := SerializeUser(user)
	if err != nil {
		return fmt.Errorf("failed to serialize user: %v", err)
	}

	// 确保目录存在
	authDir := filepath.Join(fs.basePath, "authorities")
	if err := os.MkdirAll(authDir, 0755); err != nil {
		return fmt.Errorf("failed to create authority directory: %v", err)
	}

	// 保存到文件
	userFile := filepath.Join(authDir, fmt.Sprintf("user_%s.json", authorityName))
	if err := os.WriteFile(userFile, userData, 0644); err != nil {
		return fmt.Errorf("failed to write user authority file: %v", err)
	}

	return nil
}

// LoadUserAuthority 加载用户权威机构
func (fs *FileStorage) LoadUserAuthority(authorityName string, curveParam *model.CurveParam) (*model.User, error) {
	userFile := filepath.Join(fs.basePath, "authorities", fmt.Sprintf("user_%s.json", authorityName))

	// 检查文件是否存在
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("user authority file does not exist: %s", authorityName)
	}

	// 读取文件
	data, err := os.ReadFile(userFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read user authority file: %v", err)
	}

	// 反序列化
	user, err := DeserializeUser(data, curveParam)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize user: %v", err)
	}

	return user, nil
}

// ============ 密文持久化 ============

// SaveCipher 保存密文
func (fs *FileStorage) SaveCipher(cipher *model.Cipher, fileName string, policy string, dabe *model.DABE) error {
	if cipher == nil {
		return fmt.Errorf("cipher is nil")
	}

	// 序列化密文
	cipherData, err := SerializeCipher(cipher)
	if err != nil {
		return fmt.Errorf("failed to serialize cipher: %v", err)
	}

	// 创建密文文件结构
	cipherFile := CipherFile{
		FileName:   fileName,
		Policy:     policy,
		CipherData: cipherData,
		CreateTime: getCurrentTimestamp(),
	}

	// 序列化文件结构
	fileData, err := json.Marshal(cipherFile)
	if err != nil {
		return fmt.Errorf("failed to marshal cipher file: %v", err)
	}

	// 确保目录存在
	cipherDir := filepath.Join(fs.basePath, "ciphers")
	if err := os.MkdirAll(cipherDir, 0755); err != nil {
		return fmt.Errorf("failed to create cipher directory: %v", err)
	}

	// 保存到文件
	cipherPath := filepath.Join(cipherDir, fileName+".json")
	if err := os.WriteFile(cipherPath, fileData, 0644); err != nil {
		return fmt.Errorf("failed to write cipher file: %v", err)
	}

	return nil
}

// LoadCipher 加载密文
func (fs *FileStorage) LoadCipher(fileName string, dabe *model.DABE) (*model.Cipher, error) {
	cipherPath := filepath.Join(fs.basePath, "ciphers", fileName+".json")

	// 检查文件是否存在
	if _, err := os.Stat(cipherPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("cipher file does not exist: %s", fileName)
	}

	// 读取文件
	data, err := os.ReadFile(cipherPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cipher file: %v", err)
	}

	// 反序列化文件结构
	var cipherFile CipherFile
	if err := json.Unmarshal(data, &cipherFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cipher file: %v", err)
	}

	// 反序列化密文
	cipher, err := DeserializeCipher(cipherFile.CipherData, dabe.CurveParam)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize cipher: %v", err)
	}

	return cipher, nil
}

// ============ 属性公钥持久化 ============

// SaveAttributePublicKeys 保存属性公钥集合
func (fs *FileStorage) SaveAttributePublicKeys(pkMap map[string]*model.APK) error {
	if pkMap == nil {
		return fmt.Errorf("attribute public key map is nil")
	}

	// 序列化属性公钥
	attrPKData := make(map[string][]byte)
	for attrName, apk := range pkMap {
		apkBytes, err := fs.serializer.SerializeElement(apk.Gy)
		if err != nil {
			return err
		}
		attrPKData[attrName] = apkBytes
	}

	// 保存到文件
	return fs.saveToFile("attribute_public_keys.json", attrPKData)
}

// LoadAttributePublicKeys 加载属性公钥集合
func (fs *FileStorage) LoadAttributePublicKeys(curveParam *model.CurveParam) (map[string]*model.APK, error) {
	attrPKPath := filepath.Join(fs.basePath, "attribute_public_keys.json")

	// 检查文件是否存在
	if _, err := os.Stat(attrPKPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("attribute public keys file does not exist")
	}

	// 读取文件
	data, err := os.ReadFile(attrPKPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read attribute public keys file: %v", err)
	}

	// 反序列化
	var attrPKData map[string][]byte
	if err := json.Unmarshal(data, &attrPKData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal attribute public keys: %v", err)
	}

	// 反序列化属性公钥
	pkMap := make(map[string]*model.APK)
	for attrName, apkBytes := range attrPKData {
		gy, err := fs.serializer.DeserializeElement(apkBytes, curveParam, "G1")
		if err != nil {
			return nil, err
		}
		pkMap[attrName] = &model.APK{Gy: gy}
	}

	return pkMap, nil
}

// ============ 序列化接口方法 ============

// SerializeGlobalParams 序列化全局参数为JSON
func (fs *FileStorage) SerializeGlobalParams(dabe *model.DABE) ([]byte, error) {
	return SerializeDABE(dabe)
}

// SerializeUserAuthority 序列化用户权威机构
func (fs *FileStorage) SerializeUserAuthority(user *model.User) ([]byte, error) {
	return SerializeUser(user)
}

// SerializeCipherData 序列化密文数据
func (fs *FileStorage) SerializeCipherData(cipher *model.Cipher) ([]byte, error) {
	return SerializeCipher(cipher)
}

// DeserializeUserAuthority 反序列化用户权威机构
func (fs *FileStorage) DeserializeUserAuthority(data []byte, curveParam *model.CurveParam) (*model.User, error) {
	return DeserializeUser(data, curveParam)
}

// DeserializeCipherData 反序列化密文数据
func (fs *FileStorage) DeserializeCipherData(data []byte, curveParam *model.CurveParam) (*model.Cipher, error) {
	return DeserializeCipher(data, curveParam)
}

// ============ 工具方法 ============

// saveToFile 保存数据到文件的通用方法
func (fs *FileStorage) saveToFile(relativePath string, data interface{}) error {
	// 确保目录存在
	fullPath := filepath.Join(fs.basePath, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() string {
	return time.Now().Format(time.RFC3339)
}
