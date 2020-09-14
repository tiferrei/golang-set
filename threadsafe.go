/*
Open Source Initiative OSI - The MIT License (MIT):Licensing

The MIT License (MIT)
Copyright (c) 2013 Ralph Caraveo (deckarep@gmail.com)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package mapset

import "sync"

type threadSafeSet struct {
	s threadUnsafeSet
	l sync.RWMutex
}

func newThreadSafeSet() threadSafeSet {
	return threadSafeSet{s: newThreadUnsafeSet()}
}

func (set *threadSafeSet) Add(i interface{}) bool {
	set.l.Lock()
	ret := set.s.Add(i)
	set.l.Unlock()
	return ret
}

func (set *threadSafeSet) Contains(i ...interface{}) bool {
	set.l.RLock()
	ret := set.s.Contains(i...)
	set.l.RUnlock()
	return ret
}

func (set *threadSafeSet) IsSubset(other Set) bool {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	ret := set.s.IsSubset(&o.s)
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) IsProperSubset(other Set) bool {
	o := other.(*threadSafeSet)

	set.l.RLock()
	defer set.l.RUnlock()
	o.l.RLock()
	defer o.l.RUnlock()

	return set.s.IsProperSubset(&o.s)
}

func (set *threadSafeSet) IsSuperset(other Set) bool {
	return other.IsSubset(set)
}

func (set *threadSafeSet) IsProperSuperset(other Set) bool {
	return other.IsProperSubset(set)
}

func (set *threadSafeSet) Union(other Set) Set {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	unsafeUnion := set.s.Union(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeUnion}
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) Intersect(other Set) Set {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	unsafeIntersection := set.s.Intersect(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeIntersection}
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) Difference(other Set) Set {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	unsafeDifference := set.s.Difference(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeDifference}
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) SymmetricDifference(other Set) Set {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	unsafeDifference := set.s.SymmetricDifference(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeDifference}
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) Clear() {
	set.l.Lock()
	set.s = newThreadUnsafeSet()
	set.l.Unlock()
}

func (set *threadSafeSet) Remove(i interface{}) {
	set.l.Lock()
	delete(set.s, i)
	set.l.Unlock()
}

func (set *threadSafeSet) Cardinality() int {
	set.l.RLock()
	defer set.l.RUnlock()
	return len(set.s)
}

func (set *threadSafeSet) Each(cb func(interface{}) bool) {
	set.l.RLock()
	for elem := range set.s {
		if cb(elem) {
			break
		}
	}
	set.l.RUnlock()
}

func (set *threadSafeSet) Iter() <-chan interface{} {
	ch := make(chan interface{})
	go func() {
		set.l.RLock()

		for elem := range set.s {
			ch <- elem
		}
		close(ch)
		set.l.RUnlock()
	}()

	return ch
}

func (set *threadSafeSet) Iterator() *Iterator {
	iterator, ch, stopCh := newIterator()

	go func() {
		set.l.RLock()
	L:
		for elem := range set.s {
			select {
			case <-stopCh:
				break L
			case ch <- elem:
			}
		}
		close(ch)
		set.l.RUnlock()
	}()

	return iterator
}

func (set *threadSafeSet) Equal(other Set) bool {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	ret := set.s.Equal(&o.s)
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) Clone() Set {
	set.l.RLock()

	unsafeClone := set.s.Clone().(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeClone}
	set.l.RUnlock()
	return ret
}

func (set *threadSafeSet) String() string {
	set.l.RLock()
	ret := set.s.String()
	set.l.RUnlock()
	return ret
}

func (set *threadSafeSet) PowerSet() Set {
	set.l.RLock()
	unsafePowerSet := set.s.PowerSet().(*threadUnsafeSet)
	set.l.RUnlock()

	ret := &threadSafeSet{s: newThreadUnsafeSet()}
	for subset := range unsafePowerSet.Iter() {
		unsafeSubset := subset.(*threadUnsafeSet)
		ret.Add(&threadSafeSet{s: *unsafeSubset})
	}
	return ret
}

func (set *threadSafeSet) Pop() interface{} {
	set.l.Lock()
	defer set.l.Unlock()
	return set.s.Pop()
}

func (set *threadSafeSet) CartesianProduct(other Set) Set {
	o := other.(*threadSafeSet)

	set.l.RLock()
	o.l.RLock()

	unsafeCartProduct := set.s.CartesianProduct(&o.s).(*threadUnsafeSet)
	ret := &threadSafeSet{s: *unsafeCartProduct}
	set.l.RUnlock()
	o.l.RUnlock()
	return ret
}

func (set *threadSafeSet) ToSlice() []interface{} {
	keys := make([]interface{}, 0, set.Cardinality())
	set.l.RLock()
	for elem := range set.s {
		keys = append(keys, elem)
	}
	set.l.RUnlock()
	return keys
}

func (set *threadSafeSet) MarshalJSON() ([]byte, error) {
	set.l.RLock()
	b, err := set.s.MarshalJSON()
	set.l.RUnlock()

	return b, err
}

func (set *threadSafeSet) UnmarshalJSON(p []byte) error {
	set.l.RLock()
	err := set.s.UnmarshalJSON(p)
	set.l.RUnlock()

	return err
}
