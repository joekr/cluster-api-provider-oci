package v1beta2

import (
	"context"
	"strings"
	"testing"

	"github.com/oracle/oci-go-sdk/v65/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestOCIClusterBackendSetsValidation(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add infrastructure scheme: %v", err)
	}
	if err := clusterv1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add cluster api scheme: %v", err)
	}

	newWebhook := func(objects ...ctrlclient.Object) *OCIClusterWebhook {
		builder := fake.NewClientBuilder().WithScheme(scheme)
		return &OCIClusterWebhook{Client: builder.WithObjects(objects...).Build()}
	}

	newOCICluster := func() *OCICluster {
		return &OCICluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "default",
				Labels: map[string]string{
					clusterv1.ClusterNameLabel: "owner-cluster",
				},
			},
			Spec: OCIClusterSpec{
				CompartmentId:         "ocid",
				OCIResourceIdentifier: "resource-id",
				Region:                "us-ashburn-1",
			},
		}
	}

	ownerCluster := &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "owner-cluster",
			Namespace: "default",
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: clusterv1.ClusterNetwork{
				APIServerPort: 9443,
			},
		},
	}

	t.Run("create rejects mixed legacy and canonical configuration", func(t *testing.T) {
		cluster := newOCICluster()
		cluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSetDetails: BackendSetDetails{IsFailOpen: common.Bool(true)},
			BackendSets: []BackendSet{
				{Name: "primary", ListenerPort: int32Ptr(9443)},
			},
		}

		_, err := newWebhook(ownerCluster).ValidateCreate(context.Background(), cluster)
		if err == nil || !strings.Contains(err.Error(), "legacy backendSetDetails cannot be combined with backendSets") {
			t.Fatalf("expected mixed-config validation error, got %v", err)
		}
	})

	t.Run("create rejects duplicate effective listener ports using owner cluster port", func(t *testing.T) {
		cluster := newOCICluster()
		cluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "primary"},
				{Name: "duplicate", ListenerPort: int32Ptr(9443)},
			},
		}

		_, err := newWebhook(ownerCluster).ValidateCreate(context.Background(), cluster)
		if err == nil || !strings.Contains(err.Error(), "listenerPort 9443 is duplicated") {
			t.Fatalf("expected duplicate listener port validation error, got %v", err)
		}
	})

	t.Run("update allows adding and removing secondaries order independently", func(t *testing.T) {
		oldCluster := newOCICluster()
		oldCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "primary"},
			},
		}
		newCluster := newOCICluster()
		newCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "secondary", ListenerPort: int32Ptr(7443)},
				{Name: "primary"},
			},
		}

		if _, err := newWebhook(ownerCluster).ValidateUpdate(context.Background(), oldCluster, newCluster); err != nil {
			t.Fatalf("expected secondary add to be allowed, got %v", err)
		}
		if _, err := newWebhook(ownerCluster).ValidateUpdate(context.Background(), newCluster, oldCluster); err != nil {
			t.Fatalf("expected secondary removal to be allowed, got %v", err)
		}
	})

	t.Run("update rejects renaming the primary backend set", func(t *testing.T) {
		oldCluster := newOCICluster()
		oldCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "primary"},
			},
		}
		newCluster := newOCICluster()
		newCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "renamed-primary"},
			},
		}

		_, err := newWebhook(ownerCluster).ValidateUpdate(context.Background(), oldCluster, newCluster)
		if err == nil || !strings.Contains(err.Error(), "primary backendSet ownership is immutable") {
			t.Fatalf("expected primary ownership validation error, got %v", err)
		}
	})

	t.Run("update rejects in-place listener port mutation", func(t *testing.T) {
		oldCluster := newOCICluster()
		oldCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "primary"},
				{Name: "secondary", ListenerPort: int32Ptr(7443)},
			},
		}
		newCluster := newOCICluster()
		newCluster.Spec.NetworkSpec.APIServerLB.NLBSpec = NLBSpec{
			BackendSets: []BackendSet{
				{Name: "primary"},
				{Name: "secondary", ListenerPort: int32Ptr(8443)},
			},
		}

		_, err := newWebhook(ownerCluster).ValidateUpdate(context.Background(), oldCluster, newCluster)
		if err == nil || !strings.Contains(err.Error(), "listenerPort for backendSet \"secondary\" is immutable") {
			t.Fatalf("expected immutable listenerPort validation error, got %v", err)
		}
	})
}
