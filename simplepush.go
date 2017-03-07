// Package simplepush provides a library to send (end-to-end encrypted) push messages to Smartphones via https://simplepush.io
package simplepush

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Message contains all the information necessary to send a, potentially encrypted, message.
type Message struct {
	SimplePushKey string // Your simeplepush.io key
	Title         string // Title of your message
	Message       string // Message body
	Event         string // The event the message should be associated with
	Encrypt       bool   // If set, the message will be sent end-to-end encrypted with the provided Password/Salt. If false, the message is sent unencrypted.
	Password      string // Your password
	Salt          string // If set, this salt is used, otherwise the default one gets used.
}

var defaultSalt = "1789F0B8C4A051E5"

// APIUrl is the public API entry point for https://simplepush.io. It is public to allow overriding in case
// of simplepush.io compatible services.
var APIUrl = "https://api.simplepush.io/"

func paddingPKCS5(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func encrypt(key, iv, buf []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	buf = paddingPKCS5(buf, block.BlockSize())
	if len(buf)%block.BlockSize() != 0 {
		return "", errors.New("Plaintext is not a multiple of the block size")
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(buf, buf)

	return base64.URLEncoding.EncodeToString(buf), nil
}

// Send takes a message of type Message and sends it via the simplepush.io API.
// Please refer to the documentation of the Message struct for further information.
func Send(m Message) error {
	title := m.Title
	message := m.Message
	var iv []byte

	if m.SimplePushKey == "" {
		return errors.New("simplepush.io key is not allowed to be empty")
	}

	if message == "" { // actually this is not true, but the upstream python lib handles it like that
		return errors.New("Message is not allowed to be empty")
	}

	if m.Encrypt {
		var err error
		salt := defaultSalt
		if m.Salt != "" {
			salt = m.Salt
		}
		sha := sha1.Sum([]byte(m.Password + salt))
		key, _ := hex.DecodeString(fmt.Sprintf("%X", sha)[:32])
		iv = make([]byte, aes.BlockSize)

		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return err
		}

		if title != "" {
			if title, err = encrypt(key, iv, []byte(title)); err != nil {
				return err
			}
		}

		if message, err = encrypt(key, iv, []byte(message)); err != nil {
			return err
		}
	}

	data := url.Values{}
	data.Set("key", m.SimplePushKey)
	if title != "" {
		data.Add("title", title)
	}
	data.Add("msg", message)
	data.Add("event", m.Event)
	if m.Encrypt {
		data.Add("encrypted", "true")
		data.Add("iv", fmt.Sprintf("%X", iv))
	}

	u, err := url.ParseRequestURI(APIUrl)
	if err != nil {
		return err
	}
	resource := "/send"
	u.Path = resource
	urlStr := fmt.Sprintf("%v", u)

	client := &http.Client{}

	resp, err := client.PostForm(urlStr, data)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
