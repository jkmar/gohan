package schema

import "github.com/pkg/errors"

type FilterFactory struct {
	visible, hidden []string
}

func (f *FilterFactory) CreateFilterFromProperties(visible, hidden []string) (*Filter, error) {
	f.visible = visible
	f.hidden = hidden

	strategy, err := f.createFilterStrategy()
	if err != nil {
		return nil, err
	}

	return &Filter{filterStrategy: strategy}, nil
}

func CreateExcludeAllFilter() *Filter {
	return &Filter{filterStrategy: &excludeAllFilter{}}
}

func (f *FilterFactory) createFilterStrategy() (FilterStrategy, error) {
	switch f.getFilterType() {
	case All:
		return &includeAllFilter{}, nil
	case Visible:
		return &visibleFilter{properties: sliceToSet(f.visible)}, nil
	case Hidden:
		return &hiddenFilter{properties: sliceToSet(f.hidden)}, nil
	default:
		return nil, errors.New("Cannot have filter with both visible and hidden properties")
	}
}

type filterType int

const (
	All filterType = iota
	Visible
	Hidden
	Invalid
)

func (f *FilterFactory) getFilterType() filterType {
	if f.visible == nil {
		if f.hidden == nil {
			return All
		}
		return Hidden
	}
	if f.hidden == nil {
		return Visible
	}
	return Invalid
}

func sliceToSet(keys []string) map[string]bool {
	result := make(map[string]bool)
	for _, key := range keys {
		result[key] = true
	}
	return result
}

type Filter struct {
	filterStrategy FilterStrategy
}

func (f *Filter) RemoveHiddenKeysFromMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		if f.filterStrategy.Filter(key) {
			result[key] = value
		}
	}
	return result
}

func (f *Filter) RemoveHiddenKeysFromSlice(data []string) []string {
	result := make([]string, 0)
	for _, key := range data {
		if f.filterStrategy.Filter(key) {
			result = append(result, key)
		}
	}
	return result
}

func (f *Filter) IsForbidden(key string) bool {
	return !f.filterStrategy.Filter(key)
}

func (f *Filter) AllowsAll() bool {
	_, ok := f.filterStrategy.(*includeAllFilter)
	return ok
}

type FilterStrategy interface {
	Filter(string) bool
}

type includeAllFilter struct {
}

func (i *includeAllFilter) Filter(string) bool {
	return true
}

type excludeAllFilter struct {
}

func (e *excludeAllFilter) Filter(string) bool {
	return false
}

type visibleFilter struct {
	properties map[string]bool
}

func (v *visibleFilter) Filter(s string) bool {
	return v.properties[s]
}

type hiddenFilter struct {
	properties map[string]bool
}

func (h *hiddenFilter) Filter(s string) bool {
	return !h.properties[s]
}
