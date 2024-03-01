/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Container defines container related properties.
type Container struct {
	// +kubebuilder:validation:Required
	Image string `json:"image"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`
}

// Service defines service related properties.
type Service struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port"`
}

// SchedulingConfig defines scheduling related properties.
type SchedulingConfig struct {
	// +kubebuilder:validation:Enum=Monday;Tuesday;Wednesday;Thursday;Friday;Saturday;Sunday
	StartDate string `json:"startDate"`
	// +kubebuilder:validation:Enum=Monday;Tuesday;Wednesday;Thursday;Friday;Saturday;Sunday
	EndDate string `json:"endDate"`
	// +kubebuilder:validation:Minimum=0
	Replica int `json:"replica"`
}

// ScheduledScalerSpec defines the desired state of ScheduledScaler
type ScheduledScalerSpec struct {
	// +kubebuilder:validation:Required
	Container Container `json:"container"`
	// +kubebuilder:validation:Optional
	Service Service `json:"service,omitempty"`
	// +kubebuilder:validation:Required
	SchedulingConfig *SchedulingConfig `json:"schedulingConfig"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	DefaultReplica int32 `json:"defaultReplica"`
}

// ScheduledScalerStatus defines the observed state of ScheduledScaler
type ScheduledScalerStatus struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ScheduledScaler is the Schema for the scheduledscalers API
type ScheduledScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScheduledScalerSpec   `json:"spec,omitempty"`
	Status ScheduledScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ScheduledScalerList contains a list of ScheduledScaler
type ScheduledScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScheduledScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScheduledScaler{}, &ScheduledScalerList{})
}
