# ApiTitle.DefaultApi

All URIs are relative to *http://0.0.0.0:9000*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getCamera**](DefaultApi.md#getCamera) | **GET** /api/cameras/{id} | Get all cameras
[**getCameraFiles**](DefaultApi.md#getCameraFiles) | **GET** /api/cameras/{id}/files | Get files
[**getCameraLiveStream**](DefaultApi.md#getCameraLiveStream) | **GET** /api/cameras/{id}/live | Get Live Stream
[**getCameras**](DefaultApi.md#getCameras) | **GET** /api/cameras | Get all cameras



## getCamera

> Camera getCamera(id)

Get all cameras

Get all cameras

### Example

```javascript
import ApiTitle from 'api_title';

let apiInstance = new ApiTitle.DefaultApi();
let id = 56; // Number | camera ID
apiInstance.getCamera(id, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| camera ID | 

### Return type

[**Camera**](Camera.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## getCameraFiles

> [CameraFile] getCameraFiles(id, opts)

Get files

Get camera files

### Example

```javascript
import ApiTitle from 'api_title';

let apiInstance = new ApiTitle.DefaultApi();
let id = "id_example"; // String | camera ID
let opts = {
  'start': new Date("2013-10-20T19:20:30+01:00"), // Date | range start
  'end': new Date("2013-10-20T19:20:30+01:00"), // Date | range end
  'sort': desc // String | sort order
};
apiInstance.getCameraFiles(id, opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| camera ID | 
 **start** | **Date**| range start | [optional] 
 **end** | **Date**| range end | [optional] 
 **sort** | **String**| sort order | [optional] 

### Return type

[**[CameraFile]**](CameraFile.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## getCameraLiveStream

> GetCameraLiveStream200Response getCameraLiveStream(id, opts)

Get Live Stream

Get camera live stream

### Example

```javascript
import ApiTitle from 'api_title';

let apiInstance = new ApiTitle.DefaultApi();
let id = 56; // Number | camera ID
let opts = {
  'redirect': true // Boolean | redirect request
};
apiInstance.getCameraLiveStream(id, opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **Number**| camera ID | 
 **redirect** | **Boolean**| redirect request | [optional] [default to true]

### Return type

[**GetCameraLiveStream200Response**](GetCameraLiveStream200Response.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json


## getCameras

> [Camera] getCameras(opts)

Get all cameras

Get all cameras

### Example

```javascript
import ApiTitle from 'api_title';

let apiInstance = new ApiTitle.DefaultApi();
let opts = {
  'latestSnapshot': true // Boolean | set true to return latest snapshot info
};
apiInstance.getCameras(opts, (error, data, response) => {
  if (error) {
    console.error(error);
  } else {
    console.log('API called successfully. Returned data: ' + data);
  }
});
```

### Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **latestSnapshot** | **Boolean**| set true to return latest snapshot info | [optional] 

### Return type

[**[Camera]**](Camera.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

