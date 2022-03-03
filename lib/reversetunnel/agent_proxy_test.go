package reversetunnel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgentProxyUpdates(t *testing.T) {
	proxies := NewConnectedProxies()

	expected := []string{"test"}
	proxies.updateProxyIDs(expected)

	select {
	case <-proxies.WaitForChange():
		require.Equal(t, expected, proxies.ProxyIDs())
	default:
		require.Fail(t, "Expect WaitForChange to fire.")
	}

	proxies.updateProxyIDs(append(expected, "test2"))
	proxies.updateProxyIDs(expected)

	select {
	case <-proxies.WaitForChange():
		require.Equal(t, expected, proxies.ProxyIDs())
	default:
		require.Fail(t, "Expect WaitForChange to fire.")
	}

	select {
	case <-proxies.WaitForChange():
		require.Fail(t, "Expect wait for change not to fire.")
	default:
	}

}
