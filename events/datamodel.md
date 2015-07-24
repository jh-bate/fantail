{
    "id": 
    "type": type of this event e.g. smbg
    "time": ISO8601 timestamp
    "createdAt": ISO8601 timestamp
    "updatedAt": ISO8601 timestamp
    "schemaVersion": “version” for the schema
    "data": all the data for a given type 
    "accountid": the account this data is associated 
    "deviceId":
    "uploadId":
}

type Event struct {
    //
    Id string `json:"id"`
    // The type of this event
    Type string `json:"type"`
    // An indication of the device that generated the datum. This should be globally unique to this device and repeatable with each upload.
    // A device make and model with serial number, shortened, is a good value to include here.
    DeviceId string `json:"deviceId,omitempty"`
    // The upload identifier; this field should be the uploadId of the corresponding upload data record.
    UploadId string `json:"uploadId,omitempty"`
    // An ISO8601 timestamp with a timezone offset.
    Time string `json:"time"`
    // An ISO8601 timestamp for when the items was created
    CreatedAt string `json:"createdAt,omitempty"`
    // An ISO8601 timestamp for when the item was updated
    UpdatedAt string `json:"updatedAt,omitempty"`
    //A “version” for the datum. The original datum will have a datumVersion of 0, the next modification will be 1, and so on
    DatumVersion int `json:"datumVersion"`
    //A “version” for the schema. The original schema for the type will have a schemaVersion of 0, the next modification will be 1, and so on
    SchemaVersion int `json:"schemaVersion"`
    // A flag that will indicate if the datum is valid after validation has run
    Valid bool `json:"-"`
    // The actual data for this event
    Data interface{} `json:"data"`
}