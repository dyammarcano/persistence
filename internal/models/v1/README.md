# References

- [ JSON Format v1.0 ](https://github.com/cloudevents/spec/blob/v1.0/json-format.md)


## Incomming Message Format

```json
{
    "runtimeVersion": "1.0.0",
    "dataVersion": 1,
    "metadataVersion": 1,
    "operationId": "e8e564fd-38f5-4684-9581-c30f6c25213a",
    "data": {
        "signature": "26a0a47f733d02ddb74589b6cbd6f64a7dab1947db79395a1a9e00e4c902c0f185b119897b89b248d16bab4ea781b5a3798d25c2984aec833dddab57e0891e0d68656c6c6f20776f726c64",
        "createAt": 1698763617575,
        "header": {
            "headerDate": "20230411",
            "rawHeader": "string",
            "headerLayout": "string",
            "headerTrancode": "string",
            "headerTrancodeCompl": "string",
            "structureValid": true
        },
        "credentials": {
            "familyAccount": "string"
        },
        "data": {
            "eventType": "Microsoft.Storage.BlobCreated",
            "eventTime": "2017-06-26T18:41:00.9584103Z",
            "id": "831e1650-001e-001b-66ab-eeb76e069631",
            "clientRequestId": "6d79dbfb-0e37-4fc4-981f-442c9ca65760",
            "requestId": "831e1650-001e-001b-66ab-eeb76e000000",
            "eTag": "0x8D4BCC2E4835CD0",
            "contentType": "application/octet-stream",
            "contentLength": 524288,
            "url": "https://oc2d2817345i60006.blob.core.windows.net/oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
            "sequencer": "00000000000004420000000000028963",
            "storageDiagnostics": {
                "batchId": "b68529f3-68cd-4744-baa4-3c0498ec19f0"
            }
        },
        "metadata": {
            "hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9",
            "filePath": "oc2d2817345i200097container/ArqsAguardando/oc2d2817345i20002296blob.txt",
            "outputPath": "oc2d2817345i200097container/ArqsRetorno/oc2d2817345i20002296blob.txt",
            "mimeType": "application/octet-stream",
            "sizeBytes": 524288
        },
        "origin": {
            "sftp": true
        }
    }
}
```

## Internal Message Format

```json
{
  "createAt": 1700626863623721700,                                              // extracted from persistence.Value.CreatedAt
  "updateAt": 1700626863623721700,                                              // extracted from persistence.Value.UpdatedAt
  "id": "831e1650-001e-001b-66ab-eeb76e069631",                                 // extracted from v1.IncommingEvent.IncommingData.StorageData.ID
  "operationId": "e8e564fd-38f5-4684-9581-c30f6c25213a",                        // extracted from v1.IncommingEvent.OperationID
  "runtimeVersion": "1.0.0",                                                    // extracted from v1.IncommingEvent.RuntimeVersion
  "status": "Failed",                                                           // get from system if any error or success when processing is done
  "correlationId": "831e1650-001e-001b-66ab-eeb76e000000",                      // get from system
  "fileInfo": {
    "eTag": "0x8D4BCC2E4835CD0",                                                // extracted from v1.IncommingEvent.IncommingData.StorageData.ETag
    "contentType": "application/octet-stream",                                  // extracted from v1.IncommingEvent.IncommingData.StorageData.ContentType
    "contentLength": 524288,                                                    // extracted from v1.IncommingEvent.IncommingData.StorageData.ContentLength
    "hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"  // extracted from v1.IncommingEvent.IncommingData.StorageData.Hash
  },
  "stages": {
    "solicitaionNumber": 4918754,                                               // get from system solicitation number
    "completed": "2/4",                                                         // get from system completed
    "stage1": {
      "task": "done",                                                           // get from system stage 1 task
      "startTime": "2017-06-26T18:41:00.9584103Z",                              // get from system stage 1 start time
      "endTime": "2017-06-26T18:41:00.9584103Z",                                // get from system stage 1 end time
      "message": "The request is invalid.",                                     // get from system stage 1 process message, if any and can be correlated to steps
      "innerError": {
        "date": "2017-06-26T18:41:00",                                          // get from system stage 1 process error date
        "code": "InvalidRequest",                                               // get from system stage 1 process error code
        "message": "File not meet the requirements."                            // get from system stage 1 process error message
      }
    },
    "stage2": {
      "task": "processing",                                                     // get from system stage 2 task
      "startTime": "2017-06-26T18:41:00.9584103Z",                              // get from system stage 2 start time
      "endTime": "2017-06-26T18:41:00.9584103Z",                                // get from system stage 2 end time
      "message": "The request is invalid.",                                     // get from system stage 2 process message, if any and can be correlated to steps
      "innerError": {
        "date": "2017-06-26T18:41:00",                                          // get from system stage 2 process error date
        "code": "InvalidRequest",                                               // get from system stage 2 process error code
        "message": "File not meet the requirements."                            // get from system stage 2 process error message
      }
    }
  }
}
```