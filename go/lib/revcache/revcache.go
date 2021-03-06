// Copyright 2018 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package revcache

import (
	"fmt"
	"time"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/ctrl/path_mgmt"
)

// Key denotes the key for the revocation cache.
type Key struct {
	ia   addr.IA
	ifid common.IFIDType
}

// NewKey creates a new key for the revocation cache.
func NewKey(ia addr.IA, ifid common.IFIDType) *Key {
	return &Key{
		ia:   ia,
		ifid: ifid,
	}
}

func (k Key) String() string {
	return fmt.Sprintf("%s#%s", k.ia, k.ifid)
}

// RevCache is a cache for revocations.
type RevCache interface {
	// Get item with key k from the cache. Returns the item or nil,
	// and a bool indicating whether the key was found.
	Get(k *Key) (*path_mgmt.SignedRevInfo, bool)
	// Set sets maps the key k to the revocation rev.
	// The revocation should only be returned for the given ttl.
	// If an item with key k exists, it must be updated
	// if now + ttl is at a later point in time than the current expiry.
	// Returns whether an update was performed or not.
	Set(k *Key, rev *path_mgmt.SignedRevInfo, ttl time.Duration) bool
}
