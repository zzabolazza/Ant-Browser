package backup

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	encryptedMagic     = "FACADEB1"
	encryptedChunkSize = 1024 * 1024
	encryptedSaltSize  = 16
	encryptedNonceSize = 8
)

var (
	ErrNotEncryptedBackup = errors.New("不是受支持的加密备份文件（不兼容未加密备份）")
	ErrInvalidPassword    = errors.New("密码错误或备份文件已损坏")
)

func EncryptFile(srcPath, dstPath, password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("备份密码不能为空")
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(dstPath)
		}
	}()

	header := make([]byte, 0, len(encryptedMagic)+encryptedSaltSize+encryptedNonceSize)
	header = append(header, encryptedMagic...)
	randomPart := make([]byte, encryptedSaltSize+encryptedNonceSize)
	if _, err := io.ReadFull(rand.Reader, randomPart); err != nil {
		return err
	}
	header = append(header, randomPart...)
	if _, err := out.Write(header); err != nil {
		return err
	}
	aead, err := newBackupAEAD(password, randomPart[:encryptedSaltSize])
	if err != nil {
		return err
	}
	noncePrefix := randomPart[encryptedSaltSize:]
	buf := make([]byte, encryptedChunkSize)
	for counter := uint32(0); ; counter++ {
		n, readErr := io.ReadFull(in, buf)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			return readErr
		}
		if n > 0 {
			if err := writeEncryptedRecord(out, aead, header, noncePrefix, counter, buf[:n]); err != nil {
				return err
			}
		}
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			if n > 0 {
				counter++
			}
			if err := writeEncryptedRecord(out, aead, header, noncePrefix, counter, nil); err != nil {
				return err
			}
			break
		}
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func DecryptFile(srcPath, dstPath, password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("备份密码不能为空")
	}
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	reader := bufio.NewReader(in)
	header := make([]byte, len(encryptedMagic)+encryptedSaltSize+encryptedNonceSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return ErrNotEncryptedBackup
	}
	if !bytes.Equal(header[:len(encryptedMagic)], []byte(encryptedMagic)) {
		return ErrNotEncryptedBackup
	}
	saltStart := len(encryptedMagic)
	nonceStart := saltStart + encryptedSaltSize
	aead, err := newBackupAEAD(password, header[saltStart:nonceStart])
	if err != nil {
		return err
	}
	noncePrefix := header[nonceStart:]
	out, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(dstPath)
		}
	}()
	for counter := uint32(0); ; counter++ {
		var plainLen uint32
		if err := binary.Read(reader, binary.BigEndian, &plainLen); err != nil {
			return ErrInvalidPassword
		}
		if plainLen > encryptedChunkSize {
			return ErrInvalidPassword
		}
		sealed := make([]byte, int(plainLen)+aead.Overhead())
		if _, err := io.ReadFull(reader, sealed); err != nil {
			return ErrInvalidPassword
		}
		nonce := backupNonce(noncePrefix, counter)
		aad := backupRecordAAD(header, counter, plainLen)
		plain, err := aead.Open(nil, nonce, sealed, aad)
		if err != nil {
			return ErrInvalidPassword
		}
		if plainLen == 0 {
			if _, err := reader.Peek(1); err != io.EOF {
				return ErrInvalidPassword
			}
			break
		}
		if _, err := out.Write(plain); err != nil {
			return err
		}
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func newBackupAEAD(password string, salt []byte) (cipher.AEAD, error) {
	key := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func writeEncryptedRecord(w io.Writer, aead cipher.AEAD, header, noncePrefix []byte, counter uint32, plain []byte) error {
	plainLen := uint32(len(plain))
	if err := binary.Write(w, binary.BigEndian, plainLen); err != nil {
		return err
	}
	sealed := aead.Seal(nil, backupNonce(noncePrefix, counter), plain, backupRecordAAD(header, counter, plainLen))
	_, err := w.Write(sealed)
	return err
}

func backupNonce(prefix []byte, counter uint32) []byte {
	nonce := make([]byte, 12)
	copy(nonce, prefix)
	binary.BigEndian.PutUint32(nonce[8:], counter)
	return nonce
}

func backupRecordAAD(header []byte, counter, plainLen uint32) []byte {
	aad := make([]byte, len(header)+8)
	copy(aad, header)
	binary.BigEndian.PutUint32(aad[len(header):], counter)
	binary.BigEndian.PutUint32(aad[len(header)+4:], plainLen)
	return aad
}
