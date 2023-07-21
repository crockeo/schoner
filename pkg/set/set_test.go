package set

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSet_Add(t *testing.T) {
	set := NewSet[string]()
	assert.Equal(t, true, set.Add("value"))
	assert.Equal(t, false, set.Add("value"))
	assert.Equal(t, Set[string](map[string]struct{}{"value": {}}), set)
}

func TestSet_UnionInPlace(t *testing.T) {
	set := NewSet[string]()
	set.Add("v1")

	other := NewSet[string]()
	other.Add("v2")
	other.Add("v3")

	set.UnionInPlace(other)
	slice := set.ToSlice()
	sort.Strings(slice)
	assert.Equal(t, []string{"v1", "v2", "v3"}, slice)
}

func TestSet_Remove(t *testing.T) {
	set := NewSet[string]()
	set.Add("value")
	assert.Equal(t, true, set.Remove("value"))
	assert.Equal(t, false, set.Remove("value"))
	assert.Equal(t, Set[string](map[string]struct{}{}), set)
}

func TestSet_Contains(t *testing.T) {
	set := NewSet[string]()
	assert.False(t, set.Contains("value"))
	set.Add("value")
	assert.True(t, set.Contains("value"))
}

func TestSet_Len(t *testing.T) {
	set := NewSet[string]()
	assert.Equal(t, 0, set.Len())
	set.Add("value")
	set.Add("other value")
	assert.Equal(t, 2, set.Len())
}

func TestSet_ToSlice(t *testing.T) {
	set := NewSet[string]()
	set.Add("v1")
	set.Add("v2")
	slice := set.ToSlice()
	sort.Strings(slice)
	assert.Equal(t, []string{"v1", "v2"}, slice)
}
