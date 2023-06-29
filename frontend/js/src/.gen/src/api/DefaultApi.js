/**
 * API Title
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 *
 */


import ApiClient from "../ApiClient";
import Camera from '../model/Camera';
import CameraFile from '../model/CameraFile';
import GetCameraLiveStream200Response from '../model/GetCameraLiveStream200Response';

/**
* Default service.
* @module api/DefaultApi
* @version 1.0
*/
export default class DefaultApi {

    /**
    * Constructs a new DefaultApi. 
    * @alias module:api/DefaultApi
    * @class
    * @param {module:ApiClient} [apiClient] Optional API client implementation to use,
    * default to {@link module:ApiClient#instance} if unspecified.
    */
    constructor(apiClient) {
        this.apiClient = apiClient || ApiClient.instance;
    }


    /**
     * Callback function to receive the result of the getCamera operation.
     * @callback module:api/DefaultApi~getCameraCallback
     * @param {String} error Error message, if any.
     * @param {module:model/Camera} data The data returned by the service call.
     * @param {String} response The complete HTTP response.
     */

    /**
     * Get all cameras
     * Get all cameras
     * @param {Number} id camera ID
     * @param {module:api/DefaultApi~getCameraCallback} callback The callback function, accepting three arguments: error, data, response
     * data is of type: {@link module:model/Camera}
     */
    getCamera(id, callback) {
      let postBody = null;
      // verify the required parameter 'id' is set
      if (id === undefined || id === null) {
        throw new Error("Missing the required parameter 'id' when calling getCamera");
      }

      let pathParams = {
        'id': id
      };
      let queryParams = {
      };
      let headerParams = {
      };
      let formParams = {
      };

      let authNames = [];
      let contentTypes = [];
      let accepts = ['application/json'];
      let returnType = Camera;
      return this.apiClient.callApi(
        '/api/cameras/{id}', 'GET',
        pathParams, queryParams, headerParams, formParams, postBody,
        authNames, contentTypes, accepts, returnType, null, callback
      );
    }

    /**
     * Callback function to receive the result of the getCameraFiles operation.
     * @callback module:api/DefaultApi~getCameraFilesCallback
     * @param {String} error Error message, if any.
     * @param {Array.<module:model/CameraFile>} data The data returned by the service call.
     * @param {String} response The complete HTTP response.
     */

    /**
     * Get files
     * Get camera files
     * @param {String} id camera ID
     * @param {Object} opts Optional parameters
     * @param {Date} [start] range start
     * @param {Date} [end] range end
     * @param {module:model/String} [sort] sort order
     * @param {module:api/DefaultApi~getCameraFilesCallback} callback The callback function, accepting three arguments: error, data, response
     * data is of type: {@link Array.<module:model/CameraFile>}
     */
    getCameraFiles(id, opts, callback) {
      opts = opts || {};
      let postBody = null;
      // verify the required parameter 'id' is set
      if (id === undefined || id === null) {
        throw new Error("Missing the required parameter 'id' when calling getCameraFiles");
      }

      let pathParams = {
        'id': id
      };
      let queryParams = {
        'start': opts['start'],
        'end': opts['end'],
        'sort': opts['sort']
      };
      let headerParams = {
      };
      let formParams = {
      };

      let authNames = [];
      let contentTypes = [];
      let accepts = ['application/json'];
      let returnType = [CameraFile];
      return this.apiClient.callApi(
        '/api/cameras/{id}/files', 'GET',
        pathParams, queryParams, headerParams, formParams, postBody,
        authNames, contentTypes, accepts, returnType, null, callback
      );
    }

    /**
     * Callback function to receive the result of the getCameraLiveStream operation.
     * @callback module:api/DefaultApi~getCameraLiveStreamCallback
     * @param {String} error Error message, if any.
     * @param {module:model/GetCameraLiveStream200Response} data The data returned by the service call.
     * @param {String} response The complete HTTP response.
     */

    /**
     * Get Live Stream
     * Get camera live stream
     * @param {Number} id camera ID
     * @param {Object} opts Optional parameters
     * @param {Boolean} [redirect = true)] redirect request
     * @param {module:api/DefaultApi~getCameraLiveStreamCallback} callback The callback function, accepting three arguments: error, data, response
     * data is of type: {@link module:model/GetCameraLiveStream200Response}
     */
    getCameraLiveStream(id, opts, callback) {
      opts = opts || {};
      let postBody = null;
      // verify the required parameter 'id' is set
      if (id === undefined || id === null) {
        throw new Error("Missing the required parameter 'id' when calling getCameraLiveStream");
      }

      let pathParams = {
        'id': id
      };
      let queryParams = {
        'redirect': opts['redirect']
      };
      let headerParams = {
      };
      let formParams = {
      };

      let authNames = [];
      let contentTypes = [];
      let accepts = ['application/json'];
      let returnType = GetCameraLiveStream200Response;
      return this.apiClient.callApi(
        '/api/cameras/{id}/live', 'GET',
        pathParams, queryParams, headerParams, formParams, postBody,
        authNames, contentTypes, accepts, returnType, null, callback
      );
    }

    /**
     * Callback function to receive the result of the getCameras operation.
     * @callback module:api/DefaultApi~getCamerasCallback
     * @param {String} error Error message, if any.
     * @param {Array.<module:model/Camera>} data The data returned by the service call.
     * @param {String} response The complete HTTP response.
     */

    /**
     * Get all cameras
     * Get all cameras
     * @param {Object} opts Optional parameters
     * @param {Boolean} [latestSnapshot] set true to return latest snapshot info
     * @param {module:api/DefaultApi~getCamerasCallback} callback The callback function, accepting three arguments: error, data, response
     * data is of type: {@link Array.<module:model/Camera>}
     */
    getCameras(opts, callback) {
      opts = opts || {};
      let postBody = null;

      let pathParams = {
      };
      let queryParams = {
        'latest_snapshot': opts['latestSnapshot']
      };
      let headerParams = {
      };
      let formParams = {
      };

      let authNames = [];
      let contentTypes = [];
      let accepts = ['application/json'];
      let returnType = [Camera];
      return this.apiClient.callApi(
        '/api/cameras', 'GET',
        pathParams, queryParams, headerParams, formParams, postBody,
        authNames, contentTypes, accepts, returnType, null, callback
      );
    }


}