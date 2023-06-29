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

import ApiClient from '../ApiClient';
import LatestSnapshot from './LatestSnapshot';

/**
 * The Camera model module.
 * @module model/Camera
 * @version 1.0
 */
class Camera {
    /**
     * Constructs a new <code>Camera</code>.
     * @alias module:model/Camera
     * @param id {Number} 
     * @param name {String} 
     * @param type {String} 
     * @param host {String} 
     */
    constructor(id, name, type, host) { 
        
        Camera.initialize(this, id, name, type, host);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj, id, name, type, host) { 
        obj['id'] = id;
        obj['name'] = name;
        obj['type'] = type;
        obj['host'] = host;
    }

    /**
     * Constructs a <code>Camera</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/Camera} obj Optional instance to populate.
     * @return {module:model/Camera} The populated <code>Camera</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new Camera();

            if (data.hasOwnProperty('id')) {
                obj['id'] = ApiClient.convertToType(data['id'], 'Number');
            }
            if (data.hasOwnProperty('name')) {
                obj['name'] = ApiClient.convertToType(data['name'], 'String');
            }
            if (data.hasOwnProperty('type')) {
                obj['type'] = ApiClient.convertToType(data['type'], 'String');
            }
            if (data.hasOwnProperty('host')) {
                obj['host'] = ApiClient.convertToType(data['host'], 'String');
            }
            if (data.hasOwnProperty('last_seen')) {
                obj['last_seen'] = ApiClient.convertToType(data['last_seen'], 'Date');
            }
            if (data.hasOwnProperty('enabled')) {
                obj['enabled'] = ApiClient.convertToType(data['enabled'], 'Boolean');
            }
            if (data.hasOwnProperty('timezone')) {
                obj['timezone'] = ApiClient.convertToType(data['timezone'], 'String');
            }
            if (data.hasOwnProperty('max_file_age_days')) {
                obj['max_file_age_days'] = ApiClient.convertToType(data['max_file_age_days'], 'Number');
            }
            if (data.hasOwnProperty('max_file_total_mb')) {
                obj['max_file_total_mb'] = ApiClient.convertToType(data['max_file_total_mb'], 'Number');
            }
            if (data.hasOwnProperty('username')) {
                obj['username'] = ApiClient.convertToType(data['username'], 'String');
            }
            if (data.hasOwnProperty('latest_snapshot')) {
                obj['latest_snapshot'] = LatestSnapshot.constructFromObject(data['latest_snapshot']);
            }
        }
        return obj;
    }

    /**
     * Validates the JSON data with respect to <code>Camera</code>.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @return {boolean} to indicate whether the JSON data is valid with respect to <code>Camera</code>.
     */
    static validateJSON(data) {
        // check to make sure all required properties are present in the JSON string
        for (const property of Camera.RequiredProperties) {
            if (!data[property]) {
                throw new Error("The required field `" + property + "` is not found in the JSON data: " + JSON.stringify(data));
            }
        }
        // ensure the json data is a string
        if (data['name'] && !(typeof data['name'] === 'string' || data['name'] instanceof String)) {
            throw new Error("Expected the field `name` to be a primitive type in the JSON string but got " + data['name']);
        }
        // ensure the json data is a string
        if (data['type'] && !(typeof data['type'] === 'string' || data['type'] instanceof String)) {
            throw new Error("Expected the field `type` to be a primitive type in the JSON string but got " + data['type']);
        }
        // ensure the json data is a string
        if (data['host'] && !(typeof data['host'] === 'string' || data['host'] instanceof String)) {
            throw new Error("Expected the field `host` to be a primitive type in the JSON string but got " + data['host']);
        }
        // ensure the json data is a string
        if (data['timezone'] && !(typeof data['timezone'] === 'string' || data['timezone'] instanceof String)) {
            throw new Error("Expected the field `timezone` to be a primitive type in the JSON string but got " + data['timezone']);
        }
        // ensure the json data is a string
        if (data['username'] && !(typeof data['username'] === 'string' || data['username'] instanceof String)) {
            throw new Error("Expected the field `username` to be a primitive type in the JSON string but got " + data['username']);
        }
        // validate the optional field `latest_snapshot`
        if (data['latest_snapshot']) { // data not null
          LatestSnapshot.validateJSON(data['latest_snapshot']);
        }

        return true;
    }


}

Camera.RequiredProperties = ["id", "name", "type", "host"];

/**
 * @member {Number} id
 */
Camera.prototype['id'] = undefined;

/**
 * @member {String} name
 */
Camera.prototype['name'] = undefined;

/**
 * @member {String} type
 */
Camera.prototype['type'] = undefined;

/**
 * @member {String} host
 */
Camera.prototype['host'] = undefined;

/**
 * @member {Date} last_seen
 */
Camera.prototype['last_seen'] = undefined;

/**
 * @member {Boolean} enabled
 */
Camera.prototype['enabled'] = undefined;

/**
 * @member {String} timezone
 */
Camera.prototype['timezone'] = undefined;

/**
 * @member {Number} max_file_age_days
 */
Camera.prototype['max_file_age_days'] = undefined;

/**
 * @member {Number} max_file_total_mb
 */
Camera.prototype['max_file_total_mb'] = undefined;

/**
 * @member {String} username
 */
Camera.prototype['username'] = undefined;

/**
 * @member {module:model/LatestSnapshot} latest_snapshot
 */
Camera.prototype['latest_snapshot'] = undefined;






export default Camera;
