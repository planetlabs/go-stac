package crawler

// Asset provides metadata about data for an item.
type Asset map[string]interface{}

// Href returns the asset's href.
func (a Asset) Href() string {
	value, _ := a["href"].(string)
	return value
}

// Type returns the asset's type.
func (a Asset) Type() string {
	value, _ := a["type"].(string)
	return value
}

// Title returns the asset's title.
func (a Asset) Title() string {
	value, _ := a["title"].(string)
	return value
}

// Description returns the asset's description.
func (a Asset) Description() string {
	value, _ := a["description"].(string)
	return value
}

// Roles returns the asset's description.
func (a Asset) Roles() []string {
	var roles = []string{}
	list, ok := a["roles"].([]interface{})
	if !ok {
		return roles
	}
	for _, item := range list {
		role, ok := item.(string)
		if ok {
			roles = append(roles, role)
		}
	}
	return roles
}
