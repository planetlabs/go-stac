package static

var PSScene4Band_5TO6OOBECLQ2UQBDC5JM = FeatureSchema{
	Namespace:    "/planet/imagery",
	Collection:   "PSScene4Band",
	CollectionID: "PSScene4Band_5TO6OOBECLQ2UQBDC5JM",
	Schema: `{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "id": "core.json",
  "title": "Core Asset Metadata",
  "type": "object",
  "description":
    "Core Asset Metadata",
  "additionalProperties": true,
  "required": ["id", "properties", "geometry" ],
  "allOf": [
    {
      "type": "object",
      "properties": {
        "id": { "type": "string"},
        "properties": {
          "type": "object",
          "required": ["observed"],
          "properties": {
            "_namespace": {"type": "string"},
            "_collection": {"type": "string"},
            "_updated": { "type": "string", "format": "date-time" },
            "_published": { "type": "string", "format": "date-time" },
            "_deleted": { "type": "boolean" },
            "observed": { "type": "string", "format": "date-time" },
            "duration": { "type": "number" }
          }
        }
      }
    },
    { "$ref": "#/definitions/geojson" },
    { "$ref": "#/definitions/PSScene4Band"}
  ],
  "definitions": {
      "PSScene4Band": {
        "type": "object",
        "properties": {
          "properties": {
            "type": "object",
            "required": [],
            "properties": {"acquired":{"type":"string","title":"Acquired Timestamp","description":"Timestamp representing the nominal date and time of acquisition (in UTC).","format":"date-time"},"anomalous_pixels":{"maximum":1,"minimum":0,"type":"number","title":"Anomalous Pixels","description":"percentage of anomalous pixels. Pixels that have image quality issues documented in the quality taxonomy (e.g. hot columns). This is represented spatially within the UDM.","format":"double"},"cloud_cover":{"maximum":1,"minimum":0,"type":"number","title":"Cloud Cover Ratio","description":"Ratio of the area covered by clouds to that which is uncovered.","format":"double"},"columns":{"type":"integer","title":"Columns","description":"Number of columns in the image."},"epsg_code":{"type":"integer","title":"ESPG Code","description":"The identifier for the grid cell that the imagery product is coming from if the product is an imagery tile (not used if scene)."},"ground_control":{"type":"boolean","title":"Ground Control","description":"If the image meets the positional accuracy specifications this value will be true. If the image has uncertain positional accuracy, this value will be false."},"gsd":{"type":"number","title":"GSD","description":"The ground sampling distance of the image acquisition.","format":"double"},"instrument":{"type":"string","title":"Instrument (Satellite)","description":"The generation of the satellite telescope."},"origin_x":{"type":"number","title":"Origin X","description":"ULX coordinate of the extent of the data. The coordinate references the top left corner of the top left pixel.","format":"double"},"origin_y":{"type":"number","title":"Origin Y","description":"ULY coordinate of the extent of the data. The coordinate references the top left corner of the top left pixel.","format":"double"},"pixel_resolution":{"type":"number","title":"Pixel Resolution","description":"Pixel resolution of the imagery in meters.","format":"double"},"provider":{"title":"Imagery Provider","description":"Imagery Provider, as a member of the enum of known providers."},"published":{"type":"string","title":"Published Timestamp","description":"Timestamp representing the first publish of the item id (in UTC).","format":"date-time"},"quality_category":{"title":"Quality Category","description":"Metric for image quality. To qualify for “standard” image quality an image must meet the following criteria: sun altitude greater than or equal to 10 degrees, off nadir view angle less than 20 degrees, and saturated pixels fewer than 20%. If the image does not meet these criteria it is considered “test” quality."},"rows":{"type":"integer","title":"Rows","description":"Number of rows in the image."},"satellite_id":{"type":"string","title":"Satellite ID","description":"human-readable identifier for the satellite(e.g. 0c18, RapidEye-2, Sentinel-2A)."},"strip_id":{"type":"string","title":"Strip ID","description":"Unique Identifier for the PlanetScope collect spanning the Item's capture."},"sun_azimuth":{"maximum":360,"minimum":0,"type":"number","title":"Sun Azimuth Angle","description":"Angle from true north to the sun vector projected on the horizontal plane in degrees.","format":"double"},"sun_elevation":{"maximum":90,"minimum":-90,"type":"number","title":"Sun Elevation Angle","description":"Elevation angle of the Sun in degrees.","format":"double"},"usable_data":{"maximum":1,"minimum":0,"type":"number","title":"Usable Data Ratio","description":"Ratio of the usable to unusable portion of the imagery due to cloud cover or black fill.","format":"double"},"view_angle":{"maximum":90,"minimum":-90,"type":"number","title":"Satellite View","description":"Spacecraft across-track off-nadir viewing angle used for imaging, in degrees with + being east and - being west.","format":"double"}}
          }
        }
      },
      "coordinates": {
        "title": "Coordinates",
        "type": "array",
        "items": {
          "oneOf": [{ "type": "array" }, { "type": "number" }]
        }
      },

      "geometry": {
        "title": "Geometry",
        "description":
          "A geometry is a GeoJSON object where the type member's value is one of the following strings: 'Point', 'MultiPoint', 'LineString', 'MultiLineString', 'Polygon', 'MultiPolygon', or 'GeometryCollection'.",
        "properties": {
          "type": {
            "enum": [
              "Point",
              "MultiPoint",
              "LineString",
              "MultiLineString",
              "Polygon",
              "MultiPolygon",
              "GeometryCollection"
            ]
          } }
      },

      "feature": {
        "title": "Feature",
        "description":
          "A GeoJSON object with the type 'Feature' is a feature object.\n\n* A feature object must have a member with the name 'geometry'. The value of the geometry member is a geometry object as defined above or a JSON null value.\n\n* A feature object must have a member with the name 'properties'. The value of the properties member is an object (any JSON object or a JSON null value).\n\n* If a feature has a commonly used identifier, that identifier should be included as a member of the feature object with the name 'id'.",
        "required": ["geometry", "properties"],

        "properties": {
          "type": { "enum": ["Feature"] },

          "geometry": {
            "title": "Geometry",
            "oneOf": [{ "$ref": "#/definitions/geometry" }, { "type": "null" }]
          },

          "properties": {
            "title": "Properties",
            "oneOf": [{ "type": "object" }, { "type": "null" }]
          },

          "id": {}
        }
      },

      "featurecollection": {
        "title": "Feature Collection",
        "description":
          "A GeoJSON object with the type 'FeatureCollection' is a feature collection object.\n\nAn object of type 'FeatureCollection' must have a member with the name 'features'. The value corresponding to 'features' is an array. Each element in the array is a feature object as defined above.",
        "required": ["features"],
        "properties": {
          "type": { "enum": ["FeatureCollection"] },
          "features": {
            "title": "Features",
            "type": "array",
            "items": { "$ref": "#/definitions/feature" }
          }
        }
      },

      "linearRingCoordinates": {
        "title": "Linear Ring Coordinates",
        "description":
          "A LinearRing is closed LineString with 4 or more positions. The first and last positions are equivalent (they represent equivalent points). Though a LinearRing is not explicitly represented as a GeoJSON geometry type, it is referred to in the Polygon geometry type definition.",
        "allOf": [
          { "$ref": "#/definitions/lineStringCoordinates" },
          {
            "minItems": 4
          }
        ]
      },

      "lineStringCoordinates": {
        "title": "Line String Coordinates",
        "description":
          "For type 'LineString', the 'coordinates' member must be an array of two or more positions.",
        "allOf": [
          { "$ref": "#/definitions/coordinates" },
          {
            "minLength": 2,
            "items": { "$ref": "#/definitions/position" }
          }
        ]
      },

      "polygonCoordinates": {
        "title": "Polygon Coordinates",
        "description":
          "For type 'Polygon', the 'coordinates' member must be an array of LinearRing coordinate arrays. For Polygons with multiple rings, the first must be the exterior ring and any others must be interior rings or holes.",
        "allOf": [
          { "$ref": "#/definitions/coordinates" },
          {
            "items": { "$ref": "#/definitions/linearRingCoordinates" }
          }
        ]
      },

      "position": {
        "title": "Position",
        "type": "array",
        "description":
          "A position is the fundamental geometry construct. The 'coordinates' member of a geometry object is composed of one position (in the case of a Point geometry), an array of positions (LineString or MultiPoint geometries), an array of arrays of positions (Polygons, MultiLineStrings), or a multidimensional array of positions (MultiPolygon).\n\nA position is represented by an array of numbers. There must be at least two elements, and may be more. The order of elements must follow x, y, z order (easting, northing, altitude for coordinates in a projected coordinate reference system, or longitude, latitude, altitude for coordinates in a geographic coordinate reference system). Any number of additional elements are allowed -- interpretation and meaning of additional elements is beyond the scope of this specification.",
        "minItems": 2,
        "additionalItems": true,
        "items": {
          "type": "number"
        }
      },

      "geojson": {
        "title": "GeoJSON Object",
        "type": "object",
        "description":
          "This object represents a geometry, feature, or collection of features.",
        "additionalProperties": true,
        "required": ["type"],

        "properties": {
          "type": {
            "title": "Type",
            "type": "string",
            "description": "The type of GeoJSON object.",
            "enum": [
              "Point",
              "MultiPoint",
              "LineString",
              "MultiLineString",
              "Polygon",
              "MultiPolygon",
              "GeometryCollection",
              "Feature",
              "FeatureCollection"
            ]
          },

          "crs": {
            "title": "Coordinate Reference System (CRS)",
            "description":
              "The coordinate reference system (CRS) of a GeoJSON object is determined by its 'crs' member (referred to as the CRS object below). If an object has no crs member, then its parent or grandparent object's crs member may be acquired. If no crs member can be so acquired, the default CRS shall apply to the GeoJSON object.\n\n* The default CRS is a geographic coordinate reference system, using the WGS84 datum, and with longitude and latitude units of decimal degrees.\n\n* The value of a member named 'crs' must be a JSON object (referred to as the CRS object below) or JSON null. If the value of CRS is null, no CRS can be assumed.\n\n* The crs member should be on the top-level GeoJSON object in a hierarchy (in feature collection, feature, geometry order) and should not be repeated or overridden on children or grandchildren of the object.\n\n* A non-null CRS object has two mandatory members: 'type' and 'properties'.\n\n* The value of the type member must be a string, indicating the type of CRS object.\n\n* The value of the properties member must be an object.\n\n* CRS shall not change coordinate ordering.",

            "oneOf": [
              { "type": "null" },
              {
                "type": "object",
                "required": ["type", "properties"],
                "properties": {
                  "type": {
                    "title": "CRS Type",
                    "type": "string",
                    "description":
                      "The value of the type member must be a string, indicating the type of CRS object.",
                    "minLength": 1
                  },
                  "properties": {
                    "title": "CRS Properties",
                    "type": "object"
                  }
                }
              }
            ],

            "not": {
              "anyOf": [
                {
                  "properties": {
                    "type": { "enum": ["name"] },
                    "properties": {
                      "not": {
                        "required": ["name"],
                        "properties": {
                          "name": {
                            "type": "string",
                            "minLength": 1
                          }
                        }
                      }
                    }
                  }
                },

                {
                  "properties": {
                    "type": { "enum": ["link"] },
                    "properties": {
                      "not": {
                        "title": "Link Object",
                        "type": "object",
                        "required": ["href"],
                        "properties": {
                          "href": {
                            "title": "href",
                            "type": "string",
                            "description":
                              "The value of the required 'href' member must be a dereferenceable URI.",
                            "format": "uri"
                          },

                          "type": {
                            "title": "Link Object Type",
                            "type": "string",
                            "description":
                              "The value of the optional 'type' member must be a string that hints at the format used to represent CRS parameters at the provided URI. Suggested values are: 'proj4', 'ogcwkt', 'esriwkt', but others can be used."
                          }
                        }
                      }
                    }
                  }
                }
              ]
            }
          },

          "bbox": {
            "title": "Bounding Box",
            "type": "array",
            "description":
              "To include information on the coordinate range for geometries, features, or feature collections, a GeoJSON object may have a member named 'bbox'. The value of the bbox member must be a 2*n array where n is the number of dimensions represented in the contained geometries, with the lowest values for all axes followed by the highest values. The axes order of a bbox follows the axes order of geometries. In addition, the coordinate reference system for the bbox is assumed to match the coordinate reference system of the GeoJSON object of which it is a member.",
            "minItems": 4,
            "items": {
              "type": "number"
            }
          }
        },

        "oneOf": [
          {
            "title": "Point",
            "description":
              "For type 'Point', the 'coordinates' member must be a single position.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["Point"] },
              "coordinates": {
                "allOf": [
                  { "$ref": "#/definitions/coordinates" },
                  { "$ref": "#/definitions/position" }
                ]
              }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "Multi Point Geometry",
            "description":
              "For type 'MultiPoint', the 'coordinates' member must be an array of positions.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["MultiPoint"] },
              "coordinates": {
                "allOf": [
                  { "$ref": "#/definitions/coordinates" },
                  {
                    "items": { "$ref": "#/definitions/position" }
                  }
                ] }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "Line String",
            "description":
              "For type 'LineString', the 'coordinates' member must be an array of two or more positions.\n\nA LinearRing is closed LineString with 4 or more positions. The first and last positions are equivalent (they represent equivalent points). Though a LinearRing is not explicitly represented as a GeoJSON geometry type, it is referred to in the Polygon geometry type definition.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["LineString"] },
              "coordinates": { "$ref": "#/definitions/lineStringCoordinates" }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "MultiLineString",
            "description":
              "For type 'MultiLineString', the 'coordinates' member must be an array of LineString coordinate arrays.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["MultiLineString"] },
              "coordinates": {
                "allOf": [
                  { "$ref": "#/definitions/coordinates" },
                  {
                    "items": { "$ref": "#/definitions/lineStringCoordinates" }
                  }
                ]
              }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "Polygon",
            "description":
              "For type 'Polygon', the 'coordinates' member must be an array of LinearRing coordinate arrays. For Polygons with multiple rings, the first must be the exterior ring and any others must be interior rings or holes.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["Polygon"] },
              "coordinates": { "$ref": "#/definitions/polygonCoordinates" }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "Multi-Polygon Geometry",
            "description":
              "For type 'MultiPolygon', the 'coordinates' member must be an array of Polygon coordinate arrays.",
            "required": ["coordinates"],
            "properties": {
              "type": { "enum": ["MultiPolygon"] },
              "coordinates": {
                "allOf": [
                  { "$ref": "#/definitions/coordinates" },
                  {
                    "items": { "$ref": "#/definitions/polygonCoordinates" }
                  }
                ]
              }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          {
            "title": "Geometry Collection",
            "description":
              "A GeoJSON object with type 'GeometryCollection' is a geometry object which represents a collection of geometry objects.\n\nA geometry collection must have a member with the name 'geometries'. The value corresponding to 'geometries' is an array. Each element in this array is a GeoJSON geometry object.",
            "required": ["geometries"],
            "properties": {
              "type": { "enum": ["GeometryCollection"] },
              "geometries": {
                "title": "Geometries",
                "type": "array",
                "items": { "$ref": "#/definitions/geometry" }
              }
            },
            "allOf": [{ "$ref": "#/definitions/geometry" }]
          },

          { "$ref": "#/definitions/feature" },

          { "$ref": "#/definitions/featurecollection" }
        ]
      }
    }
  }`,
}

func init() {
	err := registerSchema(PSScene4Band_5TO6OOBECLQ2UQBDC5JM)
	if err != nil {
		panic(err)
	}
}
