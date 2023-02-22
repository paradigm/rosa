/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// IMPORTANT: This file has been generated automatically, refrain from modifying it manually as all
// your changes will be lost when the file is generated again.

package v1 // github.com/openshift-online/ocm-sdk-go/osdfleetmgmt/v1

// ManagementClusterParent represents the values of the 'management_cluster_parent' type.
//
// ManagementClusterParent reference settings of the cluster.
type ManagementClusterParent struct {
	bitmap_   uint32
	clusterId string
	href      string
	kind      string
	name      string
}

// Empty returns true if the object is empty, i.e. no attribute has a value.
func (o *ManagementClusterParent) Empty() bool {
	return o == nil || o.bitmap_ == 0
}

// ClusterId returns the value of the 'cluster_id' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Parent Cluster ID
func (o *ManagementClusterParent) ClusterId() string {
	if o != nil && o.bitmap_&1 != 0 {
		return o.clusterId
	}
	return ""
}

// GetClusterId returns the value of the 'cluster_id' attribute and
// a flag indicating if the attribute has a value.
//
// Parent Cluster ID
func (o *ManagementClusterParent) GetClusterId() (value string, ok bool) {
	ok = o != nil && o.bitmap_&1 != 0
	if ok {
		value = o.clusterId
	}
	return
}

// Href returns the value of the 'href' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Reference link to internal parent cluster
func (o *ManagementClusterParent) Href() string {
	if o != nil && o.bitmap_&2 != 0 {
		return o.href
	}
	return ""
}

// GetHref returns the value of the 'href' attribute and
// a flag indicating if the attribute has a value.
//
// Reference link to internal parent cluster
func (o *ManagementClusterParent) GetHref() (value string, ok bool) {
	ok = o != nil && o.bitmap_&2 != 0
	if ok {
		value = o.href
	}
	return
}

// Kind returns the value of the 'kind' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Kind of internal parent cluster
func (o *ManagementClusterParent) Kind() string {
	if o != nil && o.bitmap_&4 != 0 {
		return o.kind
	}
	return ""
}

// GetKind returns the value of the 'kind' attribute and
// a flag indicating if the attribute has a value.
//
// Kind of internal parent cluster
func (o *ManagementClusterParent) GetKind() (value string, ok bool) {
	ok = o != nil && o.bitmap_&4 != 0
	if ok {
		value = o.kind
	}
	return
}

// Name returns the value of the 'name' attribute, or
// the zero value of the type if the attribute doesn't have a value.
//
// Parent Cluster Name
func (o *ManagementClusterParent) Name() string {
	if o != nil && o.bitmap_&8 != 0 {
		return o.name
	}
	return ""
}

// GetName returns the value of the 'name' attribute and
// a flag indicating if the attribute has a value.
//
// Parent Cluster Name
func (o *ManagementClusterParent) GetName() (value string, ok bool) {
	ok = o != nil && o.bitmap_&8 != 0
	if ok {
		value = o.name
	}
	return
}

// ManagementClusterParentListKind is the name of the type used to represent list of objects of
// type 'management_cluster_parent'.
const ManagementClusterParentListKind = "ManagementClusterParentList"

// ManagementClusterParentListLinkKind is the name of the type used to represent links to list
// of objects of type 'management_cluster_parent'.
const ManagementClusterParentListLinkKind = "ManagementClusterParentListLink"

// ManagementClusterParentNilKind is the name of the type used to nil lists of objects of
// type 'management_cluster_parent'.
const ManagementClusterParentListNilKind = "ManagementClusterParentListNil"

// ManagementClusterParentList is a list of values of the 'management_cluster_parent' type.
type ManagementClusterParentList struct {
	href  string
	link  bool
	items []*ManagementClusterParent
}

// Len returns the length of the list.
func (l *ManagementClusterParentList) Len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

// Empty returns true if the list is empty.
func (l *ManagementClusterParentList) Empty() bool {
	return l == nil || len(l.items) == 0
}

// Get returns the item of the list with the given index. If there is no item with
// that index it returns nil.
func (l *ManagementClusterParentList) Get(i int) *ManagementClusterParent {
	if l == nil || i < 0 || i >= len(l.items) {
		return nil
	}
	return l.items[i]
}

// Slice returns an slice containing the items of the list. The returned slice is a
// copy of the one used internally, so it can be modified without affecting the
// internal representation.
//
// If you don't need to modify the returned slice consider using the Each or Range
// functions, as they don't need to allocate a new slice.
func (l *ManagementClusterParentList) Slice() []*ManagementClusterParent {
	var slice []*ManagementClusterParent
	if l == nil {
		slice = make([]*ManagementClusterParent, 0)
	} else {
		slice = make([]*ManagementClusterParent, len(l.items))
		copy(slice, l.items)
	}
	return slice
}

// Each runs the given function for each item of the list, in order. If the function
// returns false the iteration stops, otherwise it continues till all the elements
// of the list have been processed.
func (l *ManagementClusterParentList) Each(f func(item *ManagementClusterParent) bool) {
	if l == nil {
		return
	}
	for _, item := range l.items {
		if !f(item) {
			break
		}
	}
}

// Range runs the given function for each index and item of the list, in order. If
// the function returns false the iteration stops, otherwise it continues till all
// the elements of the list have been processed.
func (l *ManagementClusterParentList) Range(f func(index int, item *ManagementClusterParent) bool) {
	if l == nil {
		return
	}
	for index, item := range l.items {
		if !f(index, item) {
			break
		}
	}
}
