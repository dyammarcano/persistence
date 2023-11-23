package v1

import "encoding/json"

const (
	Done       ProcessString = "done"
	Failed     ProcessString = "failed"
	Started    ProcessString = "started"
	Processing ProcessString = "processing"
)

type (
	ProcessString   string
	OperationStatus struct {
		ID             string    `json:"id,omitempty"`
		OperationID    string    `json:"operationId,omitempty"`
		RuntimeVersion string    `json:"runtimeVersion,omitempty"`
		Status         string    `json:"status,omitempty"`
		CorrelationID  string    `json:"correlationId,omitempty"`
		FileInfo       *FileInfo `json:"fileInfo,omitempty"`
		Stages         *Stages   `json:"stages,omitempty"`
	}

	FileInfo struct {
		ETag          string `json:"eTag,omitempty"`
		ContentType   string `json:"contentType,omitempty"`
		ContentLength int    `json:"contentLength,omitempty"`
		Hash          string `json:"hash,omitempty"`
	}

	Stages struct {
		SolicitationNumber int    `json:"solicitaionNumber,omitempty"`
		Completed          string `json:"completed,omitempty"`
		Stage1             *Stage `json:"stage1,omitempty"`
		Stage2             *Stage `json:"stage2,omitempty"`
		Stage3             *Stage `json:"stage3,omitempty"`
		Stage4             *Stage `json:"stage4,omitempty"`
	}

	Stage struct {
		Task       ProcessString `json:"task,omitempty"`
		StartTime  string        `json:"startTime,omitempty"`
		EndTime    string        `json:"endTime,omitempty"`
		EventTime  string        `json:"eventTime,omitempty"`
		Message    string        `json:"message,omitempty"`
		InnerError *InnerError   `json:"innerError,omitempty"`
	}

	InnerError struct {
		Date      string `json:"date,omitempty"`
		Code      string `json:"code,omitempty"`
		Message   string `json:"message,omitempty"`
		RequestId string `json:"requestId,omitempty"`
	}
)

// Serialize serializes the OperationStatus object into a JSON string.
func (o *OperationStatus) Serialize() (string, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SerializeBytes serializes the OperationStatus object into a JSON byte array.
func (o *OperationStatus) SerializeBytes() ([]byte, error) {
	data, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Deserialize deserializes the JSON string into the OperationStatus object.
func (o *OperationStatus) Deserialize(data string) error {
	return json.Unmarshal([]byte(data), o)
}
