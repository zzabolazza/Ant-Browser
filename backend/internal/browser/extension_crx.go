package browser

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// extractCRXPublicKey 从 CRX2/CRX3 包头提取用于保持商店扩展 ID 的 SPKI 公钥。
func extractCRXPublicKey(data []byte) ([]byte, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("插件包过短，无法解析 CRX 头")
	}
	if bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		return nil, fmt.Errorf("ZIP 包不含 CRX 公钥，无法保持商店扩展 ID；请从 Chrome 商店安装或使用 .crx 文件")
	}
	if !bytes.HasPrefix(data, []byte("Cr24")) {
		return nil, fmt.Errorf("不是有效的 CRX 文件")
	}

	version := binary.LittleEndian.Uint32(data[4:8])
	switch version {
	case 2:
		return extractCRX2PublicKey(data)
	case 3:
		return extractCRX3PublicKey(data)
	default:
		return nil, fmt.Errorf("不支持的 CRX 版本: %d", version)
	}
}

func extractCRX2PublicKey(data []byte) ([]byte, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("CRX2 头不完整")
	}
	pubLen := binary.LittleEndian.Uint32(data[8:12])
	sigLen := binary.LittleEndian.Uint32(data[12:16])
	if pubLen == 0 || int(16+pubLen+sigLen) > len(data) {
		return nil, fmt.Errorf("CRX2 公钥长度无效")
	}
	key := append([]byte(nil), data[16:16+pubLen]...)
	if len(key) == 0 {
		return nil, fmt.Errorf("CRX2 公钥为空")
	}
	return key, nil
}

func extractCRX3PublicKey(data []byte) ([]byte, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("CRX3 头不完整")
	}
	headerSize := binary.LittleEndian.Uint32(data[8:12])
	if headerSize == 0 || int(12+headerSize) > len(data) {
		return nil, fmt.Errorf("CRX3 头长度无效")
	}
	header := data[12 : 12+headerSize]

	var rsaKeys [][]byte
	var crxID []byte
	if err := walkProtobufBytesFields(header, func(fieldNumber int, value []byte) error {
		switch fieldNumber {
		case 2: // sha256_with_rsa
			key, err := protobufFieldBytes(value, 1)
			if err != nil {
				return err
			}
			if len(key) > 0 {
				rsaKeys = append(rsaKeys, key)
			}
		case 10000: // signed_header_data
			id, err := protobufFieldBytes(value, 1)
			if err != nil {
				return err
			}
			if len(id) == 16 {
				crxID = id
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("解析 CRX3 头失败: %w", err)
	}

	if len(rsaKeys) == 0 {
		return nil, fmt.Errorf("CRX3 头中未找到 RSA 公钥")
	}
	if len(crxID) == 16 {
		for _, key := range rsaKeys {
			sum := sha256.Sum256(key)
			if bytes.Equal(sum[:16], crxID) {
				return append([]byte(nil), key...), nil
			}
		}
	}
	// 未匹配到 crx_id 时回退第一把 RSA 公钥（常见单钥包）。
	return append([]byte(nil), rsaKeys[0]...), nil
}

func protobufFieldBytes(message []byte, wantField int) ([]byte, error) {
	var found []byte
	err := walkProtobufBytesFields(message, func(fieldNumber int, value []byte) error {
		if fieldNumber == wantField && found == nil {
			found = append([]byte(nil), value...)
		}
		return nil
	})
	return found, err
}

func walkProtobufBytesFields(message []byte, fn func(fieldNumber int, value []byte) error) error {
	i := 0
	for i < len(message) {
		tag, n := binary.Uvarint(message[i:])
		if n <= 0 {
			return fmt.Errorf("无效的 protobuf tag")
		}
		i += n
		fieldNumber := int(tag >> 3)
		wireType := int(tag & 7)
		switch wireType {
		case 0: // varint
			_, n = binary.Uvarint(message[i:])
			if n <= 0 {
				return fmt.Errorf("无效的 protobuf varint")
			}
			i += n
		case 1: // 64-bit
			if i+8 > len(message) {
				return fmt.Errorf("protobuf 64-bit 越界")
			}
			i += 8
		case 2: // length-delimited
			length, n := binary.Uvarint(message[i:])
			if n <= 0 {
				return fmt.Errorf("无效的 protobuf length")
			}
			i += n
			end := i + int(length)
			if length > uint64(len(message)-i) || end > len(message) {
				return fmt.Errorf("protobuf length 越界")
			}
			if err := fn(fieldNumber, message[i:end]); err != nil {
				return err
			}
			i = end
		case 5: // 32-bit
			if i+4 > len(message) {
				return fmt.Errorf("protobuf 32-bit 越界")
			}
			i += 4
		default:
			return fmt.Errorf("不支持的 protobuf wire type: %d", wireType)
		}
	}
	return nil
}

func extensionIDFromPublicKey(publicKey []byte) string {
	if len(publicKey) == 0 {
		return ""
	}
	sum := sha256.Sum256(publicKey)
	return encodeChromeExtensionID(sum[:16])
}

func encodeChromeExtensionID(idBytes []byte) string {
	var builder strings.Builder
	builder.Grow(len(idBytes) * 2)
	for _, b := range idBytes {
		builder.WriteByte('a' + (b >> 4))
		builder.WriteByte('a' + (b & 0x0f))
	}
	return builder.String()
}

func publicKeyToManifestKey(publicKey []byte) string {
	return base64.StdEncoding.EncodeToString(publicKey)
}

func injectManifestPublicKey(manifestData []byte, publicKey []byte) ([]byte, error) {
	if len(publicKey) == 0 {
		return nil, fmt.Errorf("公钥为空")
	}
	var raw map[string]any
	if err := json.Unmarshal(manifestData, &raw); err != nil {
		return nil, fmt.Errorf("解析 manifest.json 失败: %w", err)
	}
	raw["key"] = publicKeyToManifestKey(publicKey)
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("写入 manifest.json 公钥失败: %w", err)
	}
	return append(out, '\n'), nil
}

func writeManifestPublicKey(installDir string, publicKey []byte) ([]byte, error) {
	manifestPath := filepath.Join(installDir, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("读取已安装 manifest.json 失败: %w", err)
	}
	updated, err := injectManifestPublicKey(data, publicKey)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(manifestPath, updated, 0o644); err != nil {
		return nil, fmt.Errorf("写入 manifest.json 公钥失败: %w", err)
	}
	return updated, nil
}
