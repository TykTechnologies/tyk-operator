// +build !ignore_autogenerated

/*


Licensed under the Mozilla Public License (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.mozilla.org/en-US/MPL/2.0/

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/TykTechnologies/tyk-operator/api/model"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIDefinitionSpec) DeepCopyInto(out *APIDefinitionSpec) {
	*out = *in
	in.APIDefinitionSpec.DeepCopyInto(&out.APIDefinitionSpec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIDefinitionSpec.
func (in *APIDefinitionSpec) DeepCopy() *APIDefinitionSpec {
	if in == nil {
		return nil
	}
	out := new(APIDefinitionSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APILimit) DeepCopyInto(out *APILimit) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APILimit.
func (in *APILimit) DeepCopy() *APILimit {
	if in == nil {
		return nil
	}
	out := new(APILimit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessDefinition) DeepCopyInto(out *AccessDefinition) {
	*out = *in
	if in.Versions != nil {
		in, out := &in.Versions, &out.Versions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowedURLs != nil {
		in, out := &in.AllowedURLs, &out.AllowedURLs
		*out = make([]AccessSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessDefinition.
func (in *AccessDefinition) DeepCopy() *AccessDefinition {
	if in == nil {
		return nil
	}
	out := new(AccessDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AccessSpec) DeepCopyInto(out *AccessSpec) {
	*out = *in
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AccessSpec.
func (in *AccessSpec) DeepCopy() *AccessSpec {
	if in == nil {
		return nil
	}
	out := new(AccessSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiDefinition) DeepCopyInto(out *ApiDefinition) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiDefinition.
func (in *ApiDefinition) DeepCopy() *ApiDefinition {
	if in == nil {
		return nil
	}
	out := new(ApiDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApiDefinition) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiDefinitionList) DeepCopyInto(out *ApiDefinitionList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ApiDefinition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiDefinitionList.
func (in *ApiDefinitionList) DeepCopy() *ApiDefinitionList {
	if in == nil {
		return nil
	}
	out := new(ApiDefinitionList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ApiDefinitionList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiDefinitionStatus) DeepCopyInto(out *ApiDefinitionStatus) {
	*out = *in
	if in.LinkedByPolicies != nil {
		in, out := &in.LinkedByPolicies, &out.LinkedByPolicies
		*out = make([]model.Target, len(*in))
		copy(*out, *in)
	}
	if in.LinkedByAPIs != nil {
		in, out := &in.LinkedByAPIs, &out.LinkedByAPIs
		*out = make([]model.Target, len(*in))
		copy(*out, *in)
	}
	if in.LinkedToAPIs != nil {
		in, out := &in.LinkedToAPIs, &out.LinkedToAPIs
		*out = make([]model.Target, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiDefinitionStatus.
func (in *ApiDefinitionStatus) DeepCopy() *ApiDefinitionStatus {
	if in == nil {
		return nil
	}
	out := new(ApiDefinitionStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolicyPartitions) DeepCopyInto(out *PolicyPartitions) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolicyPartitions.
func (in *PolicyPartitions) DeepCopy() *PolicyPartitions {
	if in == nil {
		return nil
	}
	out := new(PolicyPartitions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityPolicy) DeepCopyInto(out *SecurityPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityPolicy.
func (in *SecurityPolicy) DeepCopy() *SecurityPolicy {
	if in == nil {
		return nil
	}
	out := new(SecurityPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecurityPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityPolicyList) DeepCopyInto(out *SecurityPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SecurityPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityPolicyList.
func (in *SecurityPolicyList) DeepCopy() *SecurityPolicyList {
	if in == nil {
		return nil
	}
	out := new(SecurityPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SecurityPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityPolicySpec) DeepCopyInto(out *SecurityPolicySpec) {
	*out = *in
	if in.AccessRightsArray != nil {
		in, out := &in.AccessRightsArray, &out.AccessRightsArray
		*out = make([]*AccessDefinition, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(AccessDefinition)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.AccessRights != nil {
		in, out := &in.AccessRights, &out.AccessRights
		*out = make(map[string]AccessDefinition, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Partitions != nil {
		in, out := &in.Partitions, &out.Partitions
		*out = new(PolicyPartitions)
		**out = **in
	}
	if in.MetaData != nil {
		in, out := &in.MetaData, &out.MetaData
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityPolicySpec.
func (in *SecurityPolicySpec) DeepCopy() *SecurityPolicySpec {
	if in == nil {
		return nil
	}
	out := new(SecurityPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecurityPolicyStatus) DeepCopyInto(out *SecurityPolicyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecurityPolicyStatus.
func (in *SecurityPolicyStatus) DeepCopy() *SecurityPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(SecurityPolicyStatus)
	in.DeepCopyInto(out)
	return out
}
