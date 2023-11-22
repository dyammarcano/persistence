# References

- [ JSON Format v1.0 ](https://github.com/cloudevents/spec/blob/v1.0/json-format.md)

## Internal Message Format

```json
{
    "createAt": 1700626863623721700,
    "updateAt": 1700626863623721700,
    "id": "831e1650-001e-001b-66ab-eeb76e069631",
    "operationId": "e8e564fd-38f5-4684-9581-c30f6c25213a",
    "runtimeVersion": "1.0.0",
    "status": "Failed",
    "correlationId": "831e1650-001e-001b-66ab-eeb76e000000",
    "fileInfo": {
        "startTime": "2017-06-26T18:41:00.9584103Z",
        "endTime": "2017-06-26T18:41:00.9584103Z",
        "eTag": "0x8D4BCC2E4835CD0",
        "contentType": "application/octet-stream",
        "contentLength": 524288,
        "hash": "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"
    },
    "stages": {
        "stage0": {
            "startTime": "2017-06-26T18:41:00.9584103Z",
            "endTime": "2017-06-26T18:41:00.9584103Z",
            "message": "The request is invalid.",
            "innerError": {
                "date": "2017-06-26T18:41:00",
                "code": "InvalidRequest",
                "message": "File not meet the requirements."
            }
        },
        "stage1": {
            "startTime": "2017-06-26T18:41:00.9584103Z",
            "endTime": "2017-06-26T18:41:00.9584103Z",
            "message": "The request is invalid.",
            "innerError": {
                "date": "2017-06-26T18:41:00",
                "code": "InvalidRequest",
                "message": "File not meet the requirements."
            }
        }
    }
}
```