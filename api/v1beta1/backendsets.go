package v1beta1

import "github.com/oracle/cluster-api-provider-oci/api/v1beta2"

type ResolvedAPIServerBackendSet struct {
	Name              string
	ListenerName      string
	ListenerPort      int32
	IsPrimary         bool
	BackendSetDetails BackendSetDetails
}

// HasConfiguredBackendSetDetails returns true when any legacy backend-set detail field is explicitly configured.
func HasConfiguredBackendSetDetails(details BackendSetDetails) bool {
	hubDetails := &v1beta2.BackendSetDetails{}
	_ = Convert_v1beta1_BackendSetDetails_To_v1beta2_BackendSetDetails(&details, hubDetails, nil)
	return v1beta2.HasConfiguredBackendSetDetails(*hubDetails)
}

// ResolveAPIServerNLBBackendSets returns the canonical API server backend-set topology for the given API server port.
func ResolveAPIServerNLBBackendSets(spec NLBSpec, apiServerPort int32) []ResolvedAPIServerBackendSet {
	hubSpec := &v1beta2.NLBSpec{}
	_ = Convert_v1beta1_NLBSpec_To_v1beta2_NLBSpec(&spec, hubSpec, nil)
	hubResolved := v1beta2.ResolveAPIServerNLBBackendSets(*hubSpec, apiServerPort)
	resolved := make([]ResolvedAPIServerBackendSet, 0, len(hubResolved))
	for _, backendSet := range hubResolved {
		resolved = append(resolved, ResolvedAPIServerBackendSet{
			Name:              backendSet.Name,
			ListenerName:      backendSet.ListenerName,
			ListenerPort:      backendSet.ListenerPort,
			IsPrimary:         backendSet.IsPrimary,
			BackendSetDetails: BackendSetDetails{},
		})
	}
	for i := range spec.BackendSets {
		for j := range resolved {
			if resolved[j].Name == spec.BackendSets[i].Name {
				resolved[j].BackendSetDetails = spec.BackendSets[i].BackendSetDetails
			}
		}
	}
	if len(spec.BackendSets) == 0 && len(resolved) == 1 {
		resolved[0].BackendSetDetails = spec.BackendSetDetails
	}
	return resolved
}
