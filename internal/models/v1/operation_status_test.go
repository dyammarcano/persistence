package v1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
