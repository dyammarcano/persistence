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
		ContentType        string              `json:"contentType,omitempty"`
		ContentLength      int                 `json:"contentLength,omitempty"`
		URL                string              `json:"url,omitempty"`
		Sequencer          string              `json:"sequencer,omitempty"`
		StorageDiagnostics *StorageDiagnostics `json:"storageDiagnostics,omitempty"`
	}

	PayloadData struct {
		Signature   string       `json:"signature,omitempty"`
		Timestamp   int64        `json:"timestamp,omitempty"`
		Error       bool         `json:"error,omitempty"`
		Message     string       `json:"message,omitempty"`
		Dummy       bool         `json:"dummy,omitempty"`
		Header      *Header      `json:"header,omitempty"`
		Credentials *Credentials `json:"credentials,omitempty"`
		StorageData *StorageData `json:"data,omitempty"`
		Metadata    *Metadata    `json:"metadata,omitempty"`
		Origin      *Origin      `json:"origin,omitempty"`
	}

	Credentials struct {
		FamilyAccount string `json:"familyAccount"`
	}

	Metadata struct {
		Hash       string `json:"hash,omitempty"`
		FilePath   string `json:"filePath,omitempty"`
		OutputPath string `json:"outputPath,omitempty"`
		MimeType   string `json:"mimeType,omitempty"`
		SizeBytes  int    `json:"sizeBytes,omitempty"`
	}

	Header struct {
		HeaderDate          string `json:"headerDate,omitempty"`
		RawHeader           string `json:"rawHeader,omitempty"`
		HeaderLayout        string `json:"headerLayout,omitempty"`
		HeaderTrancode      string `json:"headerTrancode,omitempty"`
		HeaderTrancodeCompl string `json:"headerTrancodeCompl,omitempty"`
		StructureValid      bool   `json:"structureValid,omitempty"`
	}

	Operation struct {
		RuntimeVersion  string          `json:"runtimeVersion,omitempty"`
		DataVersion     DataVersion     `json:"dataVersion,omitempty"`
		MetadataVersion MetadataVersion `json:"metadataVersion,omitempty"`
		OperationID     string          `json:"operationId,omitempty"`
		Encrypted       bool            `json:"encrypted,omitempty"`
		EncryptedData   *string         `json:"encryptedData,omitempty"`
		PayloadData     *PayloadData    `json:"payloadData,omitempty"`
	}

	/*
		{
			"CreateAt": 1700626863623721700,
			"UpdateAt": 1700626863623721700,
			"Payload": {
				"id": "831e1650-001e-001b-66ab-eeb76e069631",
				"operationId": "e8e564fd-38f5-4684-9581-c30f6c25213a",
				"runtimeVersion": "1.0.0",
				"startTime": "2017-06-26T18:41:00.9584103Z",
				"endTime": "2017-06-26T18:41:00.9584103Z",
				"status": "Failed",
				"correlationId": "831e1650-001e-001b-66ab-eeb76e000000",
				"fileInfo": {
					"eTag": "0x8D4BCC2E4835CD0",
					"contentType": "application/octet-stream",
					"contentLength": 524288,
					"hash":       "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
				},
				"stages": {
					"stage0": {
						"CreateAt": 1700626863623721700,
						"UpdateAt": 1700626863623721700,
						"eventTime": "2017-06-26T18:41:00.9584103Z",
						"message": "The request is invalid.",
						"innerError": {
							"date": "2017-06-26T18:41:00",
							"code": "InvalidRequest",
							"message": "The request is invalid.",
							"requestId": "831e1650-001e-001b-66ab-eeb76e000000"
						}
					}
				}
			}
		}
	*/

	OperationStatus struct {
		CreateAt       int64     `json:"createAt,omitempty"`
		UpdateAt       int64     `json:"updateAt,omitempty"`
		ID             string    `json:"id,omitempty"`
		OperationID    string    `json:"operationId,omitempty"`
		RuntimeVersion string    `json:"runtimeVersion,omitempty"`
		Status         string    `json:"status,omitempty"`
		CorrelationID  string    `json:"correlationId,omitempty"`
		FileInfo       *FileInfo `json:"fileInfo,omitempty"`
		Stages         *Stages   `json:"stages,omitempty"`
	}

	FileInfo struct {
		StartTime     string `json:"startTime,omitempty"`
		EndTime       string `json:"endTime,omitempty"`
		ETag          string `json:"eTag,omitempty"`
		ContentType   string `json:"contentType,omitempty"`
		ContentLength int    `json:"contentLength,omitempty"`
		Hash          string `json:"hash,omitempty"`
	}

	Stages struct {
		Stage0 *Stage `json:"stage0,omitempty"`
		Stage1 *Stage `json:"stage1,omitempty"`
		Stage2 *Stage `json:"stage2,omitempty"`
		Stage3 *Stage `json:"stage3,omitempty"`
		Stage4 *Stage `json:"stage4,omitempty"`
		Stage5 *Stage `json:"stage5,omitempty"`
		Stage6 *Stage `json:"stage6,omitempty"`
		Stage7 *Stage `json:"stage7,omitempty"`
	}

	Stage struct {
		StartTime  string      `json:"startTime,omitempty"`
		EndTime    string      `json:"endTime,omitempty"`
		EventTime  string      `json:"eventTime,omitempty"`
		Message    string      `json:"message,omitempty"`
		InnerError *InnerError `json:"innerError,omitempty"`
	}

	InnerError struct {
		Date      string `json:"date,omitempty"`
		Code      string `json:"code,omitempty"`
		Message   string `json:"message,omitempty"`
		RequestId string `json:"requestId,omitempty"`
	}
)

// Serialize serializes the Operation object into a JSON string.
func (o *OperationStatus) Serialize() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Deserialize deserializes the JSON string into the Operation object.
func (o *OperationStatus) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), o)
}

func (o *Origin) GetOriginKind() OriginKind {
	if o.SFTP {
		return Sftp
	}
	return ConnectDirect
}

// Serialize serializes the Operation object into a JSON string.
func (o *Operation) Serialize() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Deserialize deserializes the JSON string into the Operation object.
func (o *Operation) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), o)
}

// hashMessage hashes the message to be signed.
func (o *Operation) hashMessage() []byte {
	value := reflect.Indirect(reflect.ValueOf(o.PayloadData))
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
func (o *Operation) VerifySignature(publicKey *rsa.PublicKey) error {
	signatureBytes, err := base58.StdEncoding.DecodeString(o.PayloadData.Signature)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, o.hashMessage(), signatureBytes)
}

// SignData signs the data using the private key and returns JSON string signed.
func (o *Operation) SignData(privateKey *rsa.PrivateKey) (string, error) {
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, o.hashMessage())
	if err != nil {
		return "", err
	}
	o.PayloadData.Signature = base58.StdEncoding.EncodeToString(signature)

	return o.Serialize()
}

func (o *Operation) EncryptPayloadData(keyStr string) (string, error) {
	data, err := json.Marshal(o.PayloadData)
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
	o.PayloadData = nil

	return o.Serialize()
}

func (o *Operation) DecryptPayloadData(keyStr string) error {
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

	o.PayloadData = &PayloadData{}
	o.EncryptedData = nil

	return json.Unmarshal(decryptedData, o.PayloadData)
}
