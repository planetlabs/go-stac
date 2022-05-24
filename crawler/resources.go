package crawler

const (
	versionKey    = "stac_version"
	extensionsKey = "stac_extensions"
)

// ResourceType indicates the STAC resource type.
type ResourceType string

const (
	Item       ResourceType = "item"
	Catalog    ResourceType = "catalog"
	Collection ResourceType = "collection"
)

// Resource represents a STAC catalog, collection, or item.
type Resource map[string]interface{}

// Type returns the specific resource type.
func (r Resource) Type() ResourceType {
	value, ok := r["type"]
	if !ok {
		if _, ok := r["extent"]; ok {
			return Collection
		}
		if _, ok := r["id"]; ok {
			return Catalog
		}
		return ""
	}

	typeValue, ok := value.(string)
	if !ok {
		return ""
	}
	switch typeValue {
	case "Feature":
		return Item
	case "Catalog":
		return Catalog
	case "Collection":
		return Collection
	default:
		return ""
	}
}

// Returns the STAC / OGC Features API conformance classes (if any).
func (r Resource) ConformsTo() []string {
	value, ok := r["conformsTo"]
	if !ok {
		return nil
	}
	typeValue, ok := value.([]interface{})
	if !ok {
		return nil
	}

	conformsTo := []string{}
	for _, value := range typeValue {
		if v, ok := value.(string); ok {
			conformsTo = append(conformsTo, v)
		}
	}
	return conformsTo
}

// Returns the assets (if any).
func (r Resource) Assets() map[string]Asset {
	value, ok := r["assets"]
	if !ok {
		return nil
	}

	typeValue, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	assets := map[string]Asset{}
	for key, value := range typeValue {
		if v, ok := value.(map[string]interface{}); ok {
			assets[key] = Asset(v)
		}
	}
	return assets
}

// Version returns the STAC version.
func (r Resource) Version() string {
	value, ok := r[versionKey]
	if !ok {
		return ""
	}
	version, ok := value.(string)
	if !ok {
		return ""
	}
	return version
}

// Links returns the resource links.
func (r Resource) Links() Links {
	links := []Link{}
	value, ok := r["links"]
	if !ok {
		return links
	}
	values, ok := value.([]interface{})
	if !ok {
		return links
	}
	for _, value := range values {
		linkValue, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		link := Link{}
		for k, v := range linkValue {
			stringValue, ok := v.(string)
			if !ok {
				continue
			}
			link[k] = stringValue
		}
		links = append(links, link)
	}
	return links
}

// Extensions returns the resource extension URLs.
func (r Resource) Extensions() []string {
	extensions := []string{}
	value, ok := r[extensionsKey]
	if !ok {
		return extensions
	}
	values, ok := value.([]interface{})
	if !ok {
		return extensions
	}
	for _, value := range values {
		extension, ok := value.(string)
		if !ok {
			continue
		}
		extensions = append(extensions, extension)
	}
	return extensions
}

type featureCollectionsResponse struct {
	Collections []Resource `json:"collections"`
	Links       Links      `json:"links"`
}

type featureCollectionResponse struct {
	Features []Resource `json:"features"`
	Links    Links      `json:"links"`
}

type childrenResponse struct {
	Children []Resource `json:"children"`
	Links    Links      `json:"links"`
}
