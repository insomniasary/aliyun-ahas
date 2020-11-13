package tools

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-ahas-go-sdk/logger"
	"os"
	"path"
	"strings"
	"sync"
)

const (
	SoleilKeyName = "S"
	LuneKeyName   = "L"
	Delimiter     = "="
)

var metaFile = path.Join(GetUserHome(), ".ahas-go.meta")
var localSoleilKey = ""
var localLuneKey = ""
var mutex = sync.RWMutex{}

func GetSoleilKey() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return localSoleilKey
}

func GetLuneKey() string {
	mutex.RLock()
	defer mutex.RUnlock()
	return localLuneKey
}

func Sign(signData string) string {
	sum256 := sha256.Sum256([]byte((signData + localLuneKey)))
	encodeToString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%x", string(sum256[:]))))
	return encodeToString
}

func splitToChunk(encodeString string, size int) string {
	if len(encodeString) < size {
		return encodeString
	}
	temp := make([]string, 0, len(encodeString)/size+1)
	for len(encodeString) > 0 {
		if len(encodeString) < size {
			size = len(encodeString)
		}
		temp, encodeString = append(temp, encodeString[:size]), encodeString[size:]
	}
	return strings.Join(temp, "")
}

func Auth(sign, signData string) bool {
	expectSign := Sign(signData)
	if expectSign != sign {
		logger.Warnf("Sign not equal. ExpectSign: %s, receiveSign: %s", expectSign, sign)
		return false
	}
	return true
}

func SaveMetadataToFile(k1, k2 string) error {
	if k1 == "" || k2 == "" {
		return errors.New("SaveMetadataToFile failed: key is empty")
	}
	mutex.Lock()
	defer mutex.Unlock()
	var err error
	file, err := os.OpenFile(metaFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()
	if err != nil {
		logger.Warnf("SaveMetadataToFile failed: open file <%s> failed: %+v", metaFile, err)
		return err
	}

	_, err = file.WriteString(strings.Join([]string{SoleilKeyName, k1}, Delimiter) + "\n")
	if err != nil {
		return err
	}
	_, err = file.WriteString(strings.Join([]string{LuneKeyName, k2}, Delimiter))
	if err != nil {
		return err
	}
	localSoleilKey = k1
	localLuneKey = k2
	return nil
}

func DecryptAES(str, key string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	h := md5.New()
	_, err = h.Write([]byte(key))
	if err != nil {
		return "", err
	}
	kt := h.Sum(nil)

	block, err := aes.NewCipher(kt)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the cipherText.
	if len(cipherText) < aes.BlockSize {
		panic("cipherText too short")
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)
	return string(cipherText), nil
}
