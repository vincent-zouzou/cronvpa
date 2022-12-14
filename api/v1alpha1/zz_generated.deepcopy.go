//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/autoscaling/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	autoscaling_k8s_iov1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Condtion) DeepCopyInto(out *Condtion) {
	*out = *in
	if in.ExcludedDates != nil {
		in, out := &in.ExcludedDates, &out.ExcludedDates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.VPASpec.DeepCopyInto(&out.VPASpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Condtion.
func (in *Condtion) DeepCopy() *Condtion {
	if in == nil {
		return nil
	}
	out := new(Condtion)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronVerticalPodAutoscaler) DeepCopyInto(out *CronVerticalPodAutoscaler) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronVerticalPodAutoscaler.
func (in *CronVerticalPodAutoscaler) DeepCopy() *CronVerticalPodAutoscaler {
	if in == nil {
		return nil
	}
	out := new(CronVerticalPodAutoscaler)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CronVerticalPodAutoscaler) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronVerticalPodAutoscalerList) DeepCopyInto(out *CronVerticalPodAutoscalerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CronVerticalPodAutoscaler, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronVerticalPodAutoscalerList.
func (in *CronVerticalPodAutoscalerList) DeepCopy() *CronVerticalPodAutoscalerList {
	if in == nil {
		return nil
	}
	out := new(CronVerticalPodAutoscalerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CronVerticalPodAutoscalerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronVerticalPodAutoscalerSpec) DeepCopyInto(out *CronVerticalPodAutoscalerSpec) {
	*out = *in
	if in.Jobs != nil {
		in, out := &in.Jobs, &out.Jobs
		*out = make([]Job, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.VPASpec.DeepCopyInto(&out.VPASpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronVerticalPodAutoscalerSpec.
func (in *CronVerticalPodAutoscalerSpec) DeepCopy() *CronVerticalPodAutoscalerSpec {
	if in == nil {
		return nil
	}
	out := new(CronVerticalPodAutoscalerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronVerticalPodAutoscalerStatus) DeepCopyInto(out *CronVerticalPodAutoscalerStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]Condtion, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronVerticalPodAutoscalerStatus.
func (in *CronVerticalPodAutoscalerStatus) DeepCopy() *CronVerticalPodAutoscalerStatus {
	if in == nil {
		return nil
	}
	out := new(CronVerticalPodAutoscalerStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Job) DeepCopyInto(out *Job) {
	*out = *in
	if in.Schedules != nil {
		in, out := &in.Schedules, &out.Schedules
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.ExcludedDates != nil {
		in, out := &in.ExcludedDates, &out.ExcludedDates
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	in.VPASpec.DeepCopyInto(&out.VPASpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Job.
func (in *Job) DeepCopy() *Job {
	if in == nil {
		return nil
	}
	out := new(Job)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *VPASpec) DeepCopyInto(out *VPASpec) {
	*out = *in
	if in.TargetRef != nil {
		in, out := &in.TargetRef, &out.TargetRef
		*out = new(v1.CrossVersionObjectReference)
		**out = **in
	}
	if in.PodUpdatePolicy != nil {
		in, out := &in.PodUpdatePolicy, &out.PodUpdatePolicy
		*out = new(autoscaling_k8s_iov1.PodUpdatePolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.PodResourcePolicy != nil {
		in, out := &in.PodResourcePolicy, &out.PodResourcePolicy
		*out = new(autoscaling_k8s_iov1.PodResourcePolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.Recommenders != nil {
		in, out := &in.Recommenders, &out.Recommenders
		*out = make([]*autoscaling_k8s_iov1.VerticalPodAutoscalerRecommenderSelector, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(autoscaling_k8s_iov1.VerticalPodAutoscalerRecommenderSelector)
				**out = **in
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new VPASpec.
func (in *VPASpec) DeepCopy() *VPASpec {
	if in == nil {
		return nil
	}
	out := new(VPASpec)
	in.DeepCopyInto(out)
	return out
}
