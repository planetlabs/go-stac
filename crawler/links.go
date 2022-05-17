package crawler

import "strings"

// Link represents a link to a resource.
type Link map[string]string

// Links is a slice of links.
type Links []Link

type LinkMatcher func(link Link) bool

type linkCandidate struct {
	link     Link
	priority int
	order    int
}

func LinkTypeApplicationJSON(link Link) bool {
	return strings.ToLower(link["type"]) == "application/json"
}

func LinkTypeAnyJSON(link Link) bool {
	return strings.HasSuffix(strings.ToLower(link["type"]), "json")
}

func LinkTypeGeoJSON(link Link) bool {
	return strings.ToLower(link["type"]) == "application/geo+json"
}

func LinkTypeNone(link Link) bool {
	_, hasType := link["type"]
	return !hasType
}

func (links Links) Rel(rel string, matchers ...LinkMatcher) Link {
	candidates := []*linkCandidate{}

	for order, link := range links {
		if link["rel"] == rel {
			if len(matchers) == 0 {
				return link
			}
			for priority, matcher := range matchers {
				if matcher(link) {
					candidates = append(candidates, &linkCandidate{link: link, priority: priority, order: order})
				}
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	for _, candidate := range candidates {
		if candidate.priority > best.priority {
			continue
		}

		if candidate.priority < best.priority {
			best = candidate
			continue
		}

		if candidate.order < best.order {
			best = candidate
		}
	}

	return best.link
}
