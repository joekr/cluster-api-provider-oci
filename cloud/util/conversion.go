/*
Copyright (c) 2025 Oracle and/or its affiliates.

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

package util

import (
	"encoding/json"

	clusterv1beta1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
)

// ConvertClusterV1Beta2ToV1Beta1 converts a v1beta2 Cluster to v1beta1
// This is a temporary conversion helper during the CAPI v1.11 migration
func ConvertClusterV1Beta2ToV1Beta1(v2Cluster *clusterv1beta2.Cluster) (*clusterv1beta1.Cluster, error) {
	if v2Cluster == nil {
		return nil, nil
	}

	// Use JSON marshaling as a simple conversion method since the types are structurally compatible
	data, err := json.Marshal(v2Cluster)
	if err != nil {
		return nil, err
	}

	v1Cluster := &clusterv1beta1.Cluster{}
	if err := json.Unmarshal(data, v1Cluster); err != nil {
		return nil, err
	}

	return v1Cluster, nil
}

// ConvertMachineV1Beta2ToV1Beta1 converts a v1beta2 Machine to v1beta1
// This is a temporary conversion helper during the CAPI v1.11 migration
func ConvertMachineV1Beta2ToV1Beta1(v2Machine *clusterv1beta2.Machine) (*clusterv1beta1.Machine, error) {
	if v2Machine == nil {
		return nil, nil
	}

	// Use JSON marshaling as a simple conversion method since the types are structurally compatible
	data, err := json.Marshal(v2Machine)
	if err != nil {
		return nil, err
	}

	v1Machine := &clusterv1beta1.Machine{}
	if err := json.Unmarshal(data, v1Machine); err != nil {
		return nil, err
	}

	return v1Machine, nil
}

// ConvertMachinePoolV1Beta2ToV1Beta1 converts a v1beta2 MachinePool to v1beta1
// This is a temporary conversion helper during the CAPI v1.11 migration
func ConvertMachinePoolV1Beta2ToV1Beta1(v2MachinePool *clusterv1beta2.MachinePool) (*clusterv1beta1.MachinePool, error) {
	if v2MachinePool == nil {
		return nil, nil
	}

	// Use JSON marshaling as a simple conversion method since the types are structurally compatible
	data, err := json.Marshal(v2MachinePool)
	if err != nil {
		return nil, err
	}

	v1MachinePool := &clusterv1beta1.MachinePool{}
	if err := json.Unmarshal(data, v1MachinePool); err != nil {
		return nil, err
	}

	return v1MachinePool, nil
}
