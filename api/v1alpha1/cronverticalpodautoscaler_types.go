/*
Copyright 2022.

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

package v1alpha1

import (
	autoscaling "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronVerticalPodAutoscalerSpec defines the desired state of CronVerticalPodAutoscaler
type CronVerticalPodAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	InitialJob  string `json:"initialJob,omitempty"`
	KeepUpdated bool   `json:"keepUpdated,omitempty"`
	Jobs        []Job  `json:"jobs,omitempty"`
	VPASpec     `json:"baseVPA,omitempty" protobuf:"bytes,2,name=baseVPA"`
}

type Job struct {
	Name          string   `json:"name,omitempt"`
	Schedules     []string `json:"schedules"`
	ExcludedDates []string `json:"excludedDates,omitempty"`
	TimeZone      string   `json:"timeZone,omitempty"`
	Suspend       bool     `json:"suspend,omitempty"`
	VPASpec       `json:"patchVPA,omitempty" protobuf:"bytes,2,name=patchVPA"`
}

// type VPASpec vpav1.VerticalPodAutoscalerCheckpointSpec

type VPASpec struct {
	// +optional
	TargetRef *autoscaling.CrossVersionObjectReference `json:"targetRef" protobuf:"bytes,1,name=targetRef"`

	// +optional
	*vpav1.PodUpdatePolicy `json:"updatePolicy,omitempty" protobuf:"bytes,2,opt,name=updatePolicy"`

	// +optional
	*vpav1.PodResourcePolicy `json:"resourcePolicy,omitempty" protobuf:"bytes,3,opt,name=resourcePolicy"`

	// +optional
	Recommenders []*vpav1.VerticalPodAutoscalerRecommenderSelector `json:"recommenders,omitempty" protobuf:"bytes,4,opt,name=recommenders"`
}

// CronVerticalPodAutoscalerStatus defines the observed state of CronVerticalPodAutoscaler
type CronVerticalPodAutoscalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	CurrentJob string     `json:"currentJob,omitempty"`
	Conditions []Condtion `json:"conditions,omitempty"`
}

type Condtion struct {
	Name               string   `json:"name"`
	Schedule           string   `json:"schedule"`
	ExcludedDates      []string `json:"excludedDates,omitempty"`
	TimeZone           string   `json:"timeZone,omitempty"`
	Suspend            bool     `json:"suspend,omitempty"`
	VPASpec            `json:"patchVPA,omitempty" protobuf:"bytes,2,name=VPAspec"`
	JobID              int    `json:"jobID,omitempty"`
	CreationTime       string `json:"creationTime,omitempty"`
	NextScheduleTime   string `json:"nextScheduleTime,omitempty"`
	LastScheduleTime   string `json:"lastScheduleTime,omitempty"`
	LastSuccessfulTime string `json:"lastSuccessfulTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=cronvpa
//+kubebuilder:printcolumn:JSONPath=".status.currentJob",name=CurrentJob,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.baseVPA.targetRef.name",name=targetRef,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// CronVerticalPodAutoscaler is the Schema for the cronverticalpodautoscalers API
type CronVerticalPodAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronVerticalPodAutoscalerSpec   `json:"spec,omitempty"`
	Status CronVerticalPodAutoscalerStatus `json:"status,omitempty"`

	// Specification of the behavior of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status.
	// *vpav1.VerticalPodAutoscalerSpec `json:"template" protobuf:"bytes,2,name=template"`
}

//+kubebuilder:object:root=true

// CronVerticalPodAutoscalerList contains a list of CronVerticalPodAutoscaler
type CronVerticalPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronVerticalPodAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronVerticalPodAutoscaler{}, &CronVerticalPodAutoscalerList{})
}
