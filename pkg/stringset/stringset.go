package stringset

import "fmt"

type stringSet struct {
	set map[string]bool
}

func New(value ...string) stringSet {
	ss := stringSet{}
	ss.set = map[string]bool{}
	ss.Add(value...)
	return ss
}

func (ss stringSet) Add(value ...string) {
	for _, v := range value {
		ss.set[v] = true;
	}
}

func (ss stringSet) Contains(value string) bool {
	_, ok := ss.set[value]
	return ok;
}

func (ss stringSet) PureRemove(value string) {
	delete(ss.set, value)
}

func (ss stringSet) Remove(value string) bool {
	res := ss.Contains(value)
	if res {
		ss.PureRemove(value)
		return true
	}

	return false
}

func (ss stringSet) Len() int {
	return len(ss.set)
}

func (ss stringSet) ToSlice() []string {
	var res []string
	for k := range ss.set {
		res = append(res, k)
	}
	return res
}

func (ss stringSet) String() string {

	// if ss.Len() == 0 {
	// 	return "[]"
	// }

	// var res string
	// res = "["
	// for k := range ss.set {
	// 	res += k + " "
	// }

	// res = res[: len(res) - 1]
	// res += "]"

	// return res

	return fmt.Sprint(ss.ToSlice())
}

func (ss stringSet) IsDisjoint(other stringSet) bool {

	var setToIter *stringSet
	var setAgaints *stringSet
	if ss.Len() > other.Len() {
		setToIter = &other
		setAgaints = &ss
	} else {
		setToIter = &ss
		setAgaints = &other
	}

	for key := range setToIter.set {
		if setAgaints.Contains(key) {
			return false
		}
	}

	return true
}

// func (ss stringSet) IsSubset(other)