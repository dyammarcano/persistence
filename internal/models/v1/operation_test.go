//go:build windows

package v1

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequestMessageString(t *testing.T) {
	msgStr := "{\"dataVersion\":1,\"metadataVersion\":1,\"operationId\":\"e8e564fd-38f5-4684-9581-c30f6c25213a\",\"payloadData\":{\"signature\":\"26a0a47f733d02ddb74589b6cbd6f64a7dab1947db79395a1a9e00e4c902c0f185b119897b89b248d16bab4ea781b5a3798d25c2984aec833dddab57e0891e0d68656c6c6f20776f726c64\",\"timestamp\":1700514761,\"header\":{\"headerDate\":\"20230411\",\"rawHeader\":\"string\",\"headerLayout\":\"string\",\"headerTrancode\":\"string\",\"headerTrancodeCompl\":\"string\",\"structureValid\":true},\"credentials\":{\"familyAccount\":\"string\"},\"data\":{\"eventType\":\"Microsoft.Storage.BlobCreated\",\"eventTime\":\"2017-06-26T18:41:00.9584103Z\",\"id\":\"831e1650-001e-001b-66ab-eeb76e069631\",\"clientRequestId\":\"6d79dbfb-0e37-4fc4-981f-442c9ca65760\",\"requestId\":\"831e1650-001e-001b-66ab-eeb76e000000\",\"eTag\":\"0x8D4BCC2E4835CD0\",\"contentType\":\"application/octet-stream\",\"contentLength\":524288,\"url\":\"https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"sequencer\":\"00000000000004420000000000028963\",\"storageDiagnostics\":{\"batchId\":\"b68529f3-68cd-4744-baa4-3c0498ec19f0\"}},\"metadata\":{\"hash\":\"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9\",\"filePath\":\"oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"outputPath\":\"oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt\",\"mimeType\":\"application/octet-stream\",\"sizeBytes\":524288},\"origin\":{\"sftp\":true}}}"

	des := &Operation{}
	err := des.Deserialize(msgStr)
	assert.NoError(t, err)

	assert.Equal(t, DataVersionCurrent, des.DataVersion)
	assert.Equal(t, MetadataVersionCurrent, des.MetadataVersion)
	assert.Equal(t, "e8e564fd-38f5-4684-9581-c30f6c25213a", des.OperationID)
}

func TestSignData(t *testing.T) {
	msgStr := "{\"dataVersion\":1,\"metadataVersion\":1,\"operationId\":\"e8e564fd-38f5-4684-9581-c30f6c25213a\",\"payloadData\":{\"timestamp\":1700514761,\"header\":{\"headerDate\":\"20230411\",\"rawHeader\":\"string\",\"headerLayout\":\"string\",\"headerTrancode\":\"string\",\"headerTrancodeCompl\":\"string\",\"structureValid\":true},\"credentials\":{\"familyAccount\":\"string\"},\"data\":{\"eventType\":\"Microsoft.Storage.BlobCreated\",\"eventTime\":\"2017-06-26T18:41:00.9584103Z\",\"id\":\"831e1650-001e-001b-66ab-eeb76e069631\",\"clientRequestId\":\"6d79dbfb-0e37-4fc4-981f-442c9ca65760\",\"requestId\":\"831e1650-001e-001b-66ab-eeb76e000000\",\"eTag\":\"0x8D4BCC2E4835CD0\",\"contentType\":\"application/octet-stream\",\"contentLength\":524288,\"url\":\"https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"sequencer\":\"00000000000004420000000000028963\",\"storageDiagnostics\":{\"batchId\":\"b68529f3-68cd-4744-baa4-3c0498ec19f0\"}},\"metadata\":{\"hash\":\"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9\",\"filePath\":\"oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"outputPath\":\"oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt\",\"mimeType\":\"application/octet-stream\",\"sizeBytes\":524288},\"origin\":{\"sftp\":true}}}"

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoErrorf(t, err, "Error generating private key: %v", err)

	obj := &Operation{}
	err = obj.Deserialize(msgStr)
	assert.NoErrorf(t, err, "Error deserializing message: %v", err)

	str, err := obj.SignData(privateKey)
	assert.NoErrorf(t, err, "Error signing data: %v", err)
	assert.NotEmptyf(t, str, "Error signing data: %v", err)
}

func TestRequestMessageStringEncrypted(t *testing.T) {
	msgStr := "{\"dataVersion\":1,\"metadataVersion\":1,\"operationId\":\"e8e564fd-38f5-4684-9581-c30f6c25213a\",\"payloadData\":{\"signature\":\"26a0a47f733d02ddb74589b6cbd6f64a7dab1947db79395a1a9e00e4c902c0f185b119897b89b248d16bab4ea781b5a3798d25c2984aec833dddab57e0891e0d68656c6c6f20776f726c64\",\"timestamp\":1700514761,\"header\":{\"headerDate\":\"20230411\",\"rawHeader\":\"string\",\"headerLayout\":\"string\",\"headerTrancode\":\"string\",\"headerTrancodeCompl\":\"string\",\"structureValid\":true},\"credentials\":{\"familyAccount\":\"string\"},\"data\":{\"eventType\":\"Microsoft.Storage.BlobCreated\",\"eventTime\":\"2017-06-26T18:41:00.9584103Z\",\"id\":\"831e1650-001e-001b-66ab-eeb76e069631\",\"clientRequestId\":\"6d79dbfb-0e37-4fc4-981f-442c9ca65760\",\"requestId\":\"831e1650-001e-001b-66ab-eeb76e000000\",\"eTag\":\"0x8D4BCC2E4835CD0\",\"contentType\":\"application/octet-stream\",\"contentLength\":524288,\"url\":\"https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"sequencer\":\"00000000000004420000000000028963\",\"storageDiagnostics\":{\"batchId\":\"b68529f3-68cd-4744-baa4-3c0498ec19f0\"}},\"metadata\":{\"hash\":\"b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9\",\"filePath\":\"oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt\",\"outputPath\":\"oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt\",\"mimeType\":\"application/octet-stream\",\"sizeBytes\":524288},\"origin\":{\"sftp\":true}}}"

	des := &Operation{}
	err := des.Deserialize(msgStr)
	assert.NoError(t, err)

	assert.Equal(t, DataVersionCurrent, des.DataVersion)
	assert.Equal(t, MetadataVersionCurrent, des.MetadataVersion)
	assert.Equal(t, "e8e564fd-38f5-4684-9581-c30f6c25213a", des.OperationID)
}

func TestEncryptPayloadData(t *testing.T) {
	msgObj := &Operation{
		DataVersion:     DataVersionCurrent,
		MetadataVersion: MetadataVersionCurrent,
		OperationID:     "e8e564fd-38f5-4684-9581-c30f6c25213a",
		Encrypted:       false,
		PayloadData: &PayloadData{
			Signature: "26a0a47f733d02ddb74589b6cbd6f64a7dab1947db79395a1a9e00e4c902c0f185b119897b89b248d16bab4ea781b5a3798d25c2984aec833dddab57e0891e0d68656c6c6f20776f726c64",
			Timestamp: 1700514761,
			Error:     false,
			Message:   "",
			Dummy:     false,
			Header: &Header{
				HeaderDate:          "20230411",
				RawHeader:           "string",
				HeaderLayout:        "string",
				HeaderTrancode:      "string",
				HeaderTrancodeCompl: "string",
				StructureValid:      true,
			},
			Credentials: &Credentials{
				FamilyAccount: "string",
			},
			StorageData: &StorageData{
				EventType:       "Microsoft.Storage.BlobCreated",
				EventTime:       "2017-06-26T18:41:00.9584103Z",
				ID:              "831e1650-001e-001b-66ab-eeb76e069631",
				ClientRequestId: "6d79dbfb-0e37-4fc4-981f-442c9ca65760",
				RequestId:       "831e1650-001e-001b-66ab-eeb76e000000",
				ETag:            "0x8D4BCC2E4835CD0",
				ContentType:     "application/octet-stream",
				ContentLength:   524288,
				URL:             "https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
				Sequencer:       "00000000000004420000000000028963",
				StorageDiagnostics: &StorageDiagnostics{
					BatchId: "b68529f3-68cd-4744-baa4-3c0498ec19f0",
				},
			},
			Metadata: &Metadata{
				Hash:       "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
				FilePath:   "oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
				OutputPath: "oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt",
				MimeType:   "application/octet-stream",
				SizeBytes:  524288,
			},
			Origin: &Origin{
				SFTP:          true,
				ConnectDirect: false,
			},
		},
	}

	data, err := msgObj.EncryptPayloadData("supersecretkey")
	assert.NoErrorf(t, err, "Error encrypting data: %v", err)
	assert.NotEmptyf(t, data, "Error encrypting data: %v", err)

	obj := &Operation{}
	err = obj.Deserialize(data)
	assert.NoErrorf(t, err, "Error deserializing data: %v", err)

	err = obj.DecryptPayloadData("supersecretkey")
	assert.NoErrorf(t, err, "Error decrypting data: %v", err)

	assert.Equal(t, msgObj.DataVersion, obj.DataVersion)
}

func TestRequestMessage(t *testing.T) {
	msgObj := &Operation{
		RuntimeVersion:  "1.0.0",
		DataVersion:     DataVersionCurrent,
		MetadataVersion: MetadataVersionCurrent,
		OperationID:     "e8e564fd-38f5-4684-9581-c30f6c25213a",
		Encrypted:       false,
		PayloadData: &PayloadData{
			Signature: "26a0a47f733d02ddb74589b6cbd6f64a7dab1947db79395a1a9e00e4c902c0f185b119897b89b248d16bab4ea781b5a3798d25c2984aec833dddab57e0891e0d68656c6c6f20776f726c64",
			Timestamp: 1700514761,
			Error:     false,
			Message:   "",
			Dummy:     false,
			Header: &Header{
				HeaderDate:          "20230411",
				RawHeader:           "string",
				HeaderLayout:        "string",
				HeaderTrancode:      "string",
				HeaderTrancodeCompl: "string",
				StructureValid:      true,
			},
			Credentials: &Credentials{
				FamilyAccount: "string",
			},
			StorageData: &StorageData{
				EventType:       "Microsoft.Storage.BlobCreated",
				EventTime:       "2017-06-26T18:41:00.9584103Z",
				ID:              "831e1650-001e-001b-66ab-eeb76e069631",
				ClientRequestId: "6d79dbfb-0e37-4fc4-981f-442c9ca65760",
				RequestId:       "831e1650-001e-001b-66ab-eeb76e000000",
				ETag:            "0x8D4BCC2E4835CD0",
				ContentType:     "application/octet-stream",
				ContentLength:   524288,
				URL:             "https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
				Sequencer:       "00000000000004420000000000028963",
				StorageDiagnostics: &StorageDiagnostics{
					BatchId: "b68529f3-68cd-4744-baa4-3c0498ec19f0",
				},
			},
			Metadata: &Metadata{
				Hash:       "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
				FilePath:   "oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
				OutputPath: "oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt",
				MimeType:   "application/octet-stream",
				SizeBytes:  524288,
			},
			Origin: &Origin{
				SFTP:          true,
				ConnectDirect: false,
			},
		},
	}

	msg, err := msgObj.Serialize()
	assert.NoError(t, err)

	des := &Operation{}
	err = des.Deserialize(msg)
	assert.NoError(t, err)

	assert.Equal(t, msgObj.DataVersion, des.DataVersion)
}

func TestOperationStatus(t *testing.T) {
	msgObj := &OperationStatus{
		RuntimeVersion: "1.0.0",
		CreateAt:       1700626863623721700,
		UpdateAt:       1700626863623721700,
		ID:             "831e1650-001e-001b-66ab-eeb76e069631",
		OperationID:    "e8e564fd-38f5-4684-9581-c30f6c25213a",
		Status:         "Failed",
		CorrelationID:  "831e1650-001e-001b-66ab-eeb76e000000",
		FileInfo: &FileInfo{
			StartTime:     "2017-06-26T18:41:00.9584103Z",
			EndTime:       "2017-06-26T18:41:00.9584103Z",
			ETag:          "0x8D4BCC2E4835CD0",
			ContentType:   "application/octet-stream",
			ContentLength: 524288,
			Hash:          "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
		},
		Stages: &Stages{
			Stage0: &Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
			Stage1: &Stage{
				StartTime: "2017-06-26T18:41:00.9584103Z",
				EndTime:   "2017-06-26T18:41:00.9584103Z",
				Message:   "The request is invalid.",
				InnerError: &InnerError{
					Date:    "2017-06-26T18:41:00",
					Code:    "InvalidRequest",
					Message: "File not meet the requirements.",
				},
			},
		},
	}

	msg, err := msgObj.Serialize()
	assert.NoError(t, err)

	des := &OperationStatus{}
	err = des.Deserialize(msg)
	assert.NoError(t, err)

	assert.Equal(t, msgObj.FileInfo.ETag, des.FileInfo.ETag)
}
