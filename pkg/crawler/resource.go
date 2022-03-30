package crawler

const (
	versionKey    = "stac_version"
	extensionsKey = "stac_extensions"
)

type ResourceType string

const (
	Item       = ResourceType("item")
	Catalog    = ResourceType("catalog")
	Collection = ResourceType("collection")
)

type Resource map[string]interface{}

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

type Link map[string]string

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
