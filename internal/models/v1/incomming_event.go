package v1

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/dyammarcano/base58"
	"reflect"
)

const (
	MetadataVersionOld MetadataVersion = iota
	MetadataVersionCurrent
	MetadataVersionExperimental
)

const (
	DataVersionOld DataVersion = iota
	DataVersionCurrent
	DataVersionExperimental
)

const (
	Sftp OriginKind = iota
	ConnectDirect
)

type (
	DataVersion     int
	MetadataVersion int
	OriginKind      int

	Origin struct {
		SFTP          bool `json:"sftp,omitempty"`
		ConnectDirect bool `json:"connectDirect,omitempty"`
	}

	StorageDiagnostics struct {
		BatchId string `json:"batchId,omitempty"`
	}

	StorageData struct {
		EventType          string              `json:"eventType,omitempty"`
		EventTime          string              `json:"eventTime,omitempty"`
		ID                 string              `json:"id,omitempty"`
		ClientRequestId    string              `json:"clientRequestId,omitempty"`
		RequestId          string              `json:"requestId,omitempty"`
		ETag               string              `json:"eTag,omitempty"`
		Hash               string              `json:"hash,omitempty"`
		ContentType        string              `json:"contentType,omitempty"`
		ContentLength      int                 `json:"contentLength,omitempty"`
		URL                string              `json:"url,omitempty"`
		Sequencer          string              `json:"sequencer,omitempty"`
		StorageDiagnostics *StorageDiagnostics `json:"storageDiagnostics,omitempty"`
	}

	IncommingData struct {
		Signature     string         `json:"signature,omitempty"`
		CreateAt      int64          `json:"createAt,omitempty"`
		Error         bool           `json:"error,omitempty"`
		Message       string         `json:"message,omitempty"`
		Dummy         bool           `json:"dummy,omitempty"`
		FileHeader    *FileHeader    `json:"header,omitempty"`
		InvokedClient *InvokedClient `json:"invokedClient,omitempty"`
		StorageData   *StorageData   `json:"data,omitempty"`
		Metadata      *Metadata      `json:"metadata,omitempty"`
		Origin        *Origin        `json:"origin,omitempty"`
	}

	InvokedClient struct {
		FamilyAccount string `json:"familyAccount"`
	}

	Metadata struct {
		FilePath   string `json:"filePath,omitempty"`
		OutputPath string `json:"outputPath,omitempty"`
		MimeType   string `json:"mimeType,omitempty"`
		SizeBytes  int    `json:"sizeBytes,omitempty"`
	}

	FileHeader struct {
		HeaderDate          string `json:"headerDate,omitempty"`
		RawHeader           string `json:"rawHeader,omitempty"`
		HeaderLayout        string `json:"headerLayout,omitempty"`
		HeaderTrancode      string `json:"headerTrancode,omitempty"`
		HeaderTrancodeCompl string `json:"headerTrancodeCompl,omitempty"`
		StructureValid      bool   `json:"structureValid,omitempty"`
	}

	IncommingEvent struct {
		RuntimeVersion  string          `json:"runtimeVersion,omitempty"`
		DataVersion     DataVersion     `json:"dataVersion,omitempty"`
		MetadataVersion MetadataVersion `json:"metadataVersion,omitempty"`
		OperationID     string          `json:"operationId,omitempty"`
		Encrypted       bool            `json:"encrypted,omitempty"`
		EncryptedData   *string         `json:"encryptedData,omitempty"`
		IncommingData   *IncommingData  `json:"incommingData,omitempty"`
	}
)

func (o *Origin) GetOriginKind() OriginKind {
	if o.SFTP {
		return Sftp
	}
	return ConnectDirect
}

// Serialize serializes the IncommingEvent object into a JSON string.
func (o *IncommingEvent) Serialize() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Deserialize deserializes the JSON string into the IncommingEvent object.
func (o *IncommingEvent) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), o)
}

// hashMessage hashes the message to be signed.
func (o *IncommingEvent) hashMessage() []byte {
	value := reflect.Indirect(reflect.ValueOf(o.IncommingData))
	typeOfValue := value.Type()

	var buffer bytes.Buffer
	for i := 0; i < value.NumField(); i++ {
		fieldName := typeOfValue.Field(i).Name
		fieldValue := value.Field(i).Interface()
		buffer.WriteString(fmt.Sprintf("%s:%v", fieldName, fieldValue))
	}

	buffer.WriteString(fmt.Sprintf("%s:%v", "operationId", o.OperationID))

	hash := sha256.Sum256(buffer.Bytes())
	return hash[:]
}

// VerifySignature verifies the signature of the data using the public key.
func (o *IncommingEvent) VerifySignature(publicKey *rsa.PublicKey) error {
	signatureBytes, err := base58.StdEncoding.DecodeString(o.IncommingData.Signature)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, o.hashMessage(), signatureBytes)
}

// SignData signs the data using the private key and returns JSON string signed.
func (o *IncommingEvent) SignData(privateKey *rsa.PrivateKey) (string, error) {
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, o.hashMessage())
	if err != nil {
		return "", err
	}
	o.IncommingData.Signature = base58.StdEncoding.EncodeToString(signature)

	return o.Serialize()
}

func (o *IncommingEvent) EncryptPayloadData(keyStr string) (string, error) {
	data, err := json.Marshal(o.IncommingData)
	if err != nil {
		return "", err
	}

	var key bytes.Buffer
	key.WriteString(fmt.Sprintf("%s", sha256.Sum256([]byte(keyStr))))

	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12) // 96-bit nonce for GCM
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, data, nil)
	encryptedData := base58.StdEncoding.EncodeToString(append(nonce, ciphertext...))

	o.Encrypted = true
	o.EncryptedData = &encryptedData
	o.IncommingData = nil

	return o.Serialize()
}

func (o *IncommingEvent) DecryptPayloadData(keyStr string) error {
	encryptedData, err := base58.StdEncoding.DecodeString(*o.EncryptedData)
	if err != nil {
		return err
	}

	var key bytes.Buffer
	key.WriteString(fmt.Sprintf("%s", sha256.Sum256([]byte(keyStr))))

	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce, ciphertext := encryptedData[:12], encryptedData[12:]
	decryptedData, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return err
	}

	o.IncommingData = &IncommingData{}
	o.EncryptedData = nil

	return json.Unmarshal(decryptedData, o.IncommingData)
}
