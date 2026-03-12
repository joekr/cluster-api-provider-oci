package v1beta2

import (
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
)

func int32Ptr(v int32) *int32 {
	return &v
}

func TestResolveAPIServerNLBBackendSets(t *testing.T) {
	t.Run("legacy fallback uses default backend set", func(t *testing.T) {
		resolved := ResolveAPIServerNLBBackendSets(NLBSpec{
			BackendSetDetails: BackendSetDetails{
				HealthChecker: HealthChecker{UrlPath: common.String("/readyz")},
			},
		}, 6443)

		if len(resolved) != 1 {
			t.Fatalf("expected 1 backend set, got %d", len(resolved))
		}
		if resolved[0].Name != APIServerLBBackendSetName {
			t.Fatalf("expected default backend set name, got %q", resolved[0].Name)
		}
		if resolved[0].ListenerName != APIServerLBListener {
			t.Fatalf("expected default listener name, got %q", resolved[0].ListenerName)
		}
		if resolved[0].ListenerPort != 6443 {
			t.Fatalf("expected default listener port 6443, got %d", resolved[0].ListenerPort)
		}
		if !resolved[0].IsPrimary {
			t.Fatalf("expected legacy fallback backend set to be primary")
		}
	})

	t.Run("primary ownership is derived by effective API server port", func(t *testing.T) {
		resolved := ResolveAPIServerNLBBackendSets(NLBSpec{
			BackendSets: []BackendSet{
				{Name: "secondary", ListenerPort: int32Ptr(7443)},
				{Name: "primary"},
			},
		}, 6443)

		if len(resolved) != 2 {
			t.Fatalf("expected 2 backend sets, got %d", len(resolved))
		}
		if resolved[0].Name != "primary" || !resolved[0].IsPrimary {
			t.Fatalf("expected primary backend set first, got %+v", resolved[0])
		}
		if resolved[0].ListenerName != APIServerLBListener {
			t.Fatalf("expected primary listener name %q, got %q", APIServerLBListener, resolved[0].ListenerName)
		}
		if resolved[1].ListenerName != APIServerLBListener+"-secondary" {
			t.Fatalf("expected deterministic secondary listener name, got %q", resolved[1].ListenerName)
		}
		if resolved[1].ListenerPort != 7443 {
			t.Fatalf("expected secondary listener port 7443, got %d", resolved[1].ListenerPort)
		}
	})
}
