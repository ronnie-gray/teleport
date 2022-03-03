/*
Copyright 2022 Gravitational, Inc.

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

package reversetunnel

import (
	"strings"
	"sync"

	"github.com/google/go-cmp/cmp"
)

// NewConnectedProxies creates a new ConnectedProxies instance.
func NewConnectedProxies() *ConnectedProxies {
	return &ConnectedProxies{
		change: make(chan struct{}, 1),
	}
}

// ConnectedProxies signals when the proxies an agent is connected to changes.
// When proxy peering is not enabled the connected proxies should be an empty list.
type ConnectedProxies struct {
	ids    []string
	change chan struct{}
	mu     sync.RWMutex
}

// ProxyIDs gets the list of proxies the agent is connected to.
func (p *ConnectedProxies) ProxyIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.ids
}

// WaitForChange is signaled when the list of connected proxies changes.
func (p *ConnectedProxies) WaitForChange() <-chan struct{} {
	return p.change
}

// updateProxyIDs updates the proxy ids and signals a change if necessary.
func (p *ConnectedProxies) updateProxyIDs(ids []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cmp.Equal(p.ids, ids) {
		return
	}

	p.ids = ids

	// Multiple changes are compacted into a single signal if sending
	// is blocked.
	select {
	case p.change <- struct{}{}:
	default:
	}
}

// getIDFromPrincipals gets the proxy id from a list of principals.
func getIDFromPrincipals(principals []string) (string, bool) {
	if len(principals) == 0 {
		return "", false
	}

	// ID will always be the first principal.
	id := principals[0]

	// Return the uuid from the format "<uuid>.<cluster-name>".
	if split := strings.Split(id, "."); len(split) > 1 {
		id = split[0]
	}

	return id, true
}
