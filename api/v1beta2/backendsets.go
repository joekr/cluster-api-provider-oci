package v1beta2

import (
	"fmt"
	"sort"
)

// ResolvedAPIServerBackendSet is the canonical API server backend-set topology used by admission and runtime.
type ResolvedAPIServerBackendSet struct {
	Name              string
	ListenerName      string
	ListenerPort      int32
	IsPrimary         bool
	BackendSetDetails BackendSetDetails
}

// HasConfiguredBackendSetDetails returns true when any legacy backend-set detail field is explicitly configured.
func HasConfiguredBackendSetDetails(details BackendSetDetails) bool {
	return details.IsPreserveSource != nil ||
		details.IsFailOpen != nil ||
		details.IsInstantFailoverEnabled != nil ||
		details.HealthChecker.UrlPath != nil
}

// ResolveAPIServerNLBBackendSets returns the canonical API server backend-set topology for the given API server port.
func ResolveAPIServerNLBBackendSets(spec NLBSpec, apiServerPort int32) []ResolvedAPIServerBackendSet {
	if len(spec.BackendSets) == 0 {
		return []ResolvedAPIServerBackendSet{
			{
				Name:              APIServerLBBackendSetName,
				ListenerName:      APIServerLBListener,
				ListenerPort:      apiServerPort,
				IsPrimary:         true,
				BackendSetDetails: spec.BackendSetDetails,
			},
		}
	}

	resolved := make([]ResolvedAPIServerBackendSet, 0, len(spec.BackendSets))
	for _, backendSet := range spec.BackendSets {
		listenerPort := apiServerPort
		if backendSet.ListenerPort != nil {
			listenerPort = *backendSet.ListenerPort
		}

		isPrimary := listenerPort == apiServerPort
		resolved = append(resolved, ResolvedAPIServerBackendSet{
			Name:              backendSet.Name,
			ListenerName:      ListenerNameForBackendSet(backendSet.Name, isPrimary),
			ListenerPort:      listenerPort,
			IsPrimary:         isPrimary,
			BackendSetDetails: backendSet.BackendSetDetails,
		})
	}

	sort.Slice(resolved, func(i, j int) bool {
		if resolved[i].IsPrimary != resolved[j].IsPrimary {
			return resolved[i].IsPrimary
		}
		return resolved[i].Name < resolved[j].Name
	})

	return resolved
}

// ListenerNameForBackendSet returns the deterministic NLB listener name for the given backend set.
func ListenerNameForBackendSet(name string, isPrimary bool) string {
	if isPrimary {
		return APIServerLBListener
	}
	return fmt.Sprintf("%s-%s", APIServerLBListener, name)
}
