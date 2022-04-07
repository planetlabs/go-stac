package crawler

const (
	versionKey    = "stac_version"
	extensionsKey = "stac_extensions"
)

// ResourceType indicates the document-level STAC resource type.
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

// Link represents a link to a resource.
type Link map[string]string

// Links returns the resource links.
func (r Resource) Links() []Link {
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
