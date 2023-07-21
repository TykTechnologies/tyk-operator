# Understanding Reconciliation Status

From [Tyk Operator v0.15.0](https://github.com/TykTechnologies/tyk-operator/releases/tag/v0.15.0), 
we introduce a new status subresource in APIDefinition CRD, called _latestTransaction_ which holds information about 
latest reconciliation status.

The [Status subresource]
(https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#status-subresource) 
in Kubernetes is a specialized endpoint that allows developers and operators to retrieve the real-time status of a 
specific Kubernetes resource. By querying this subresource, users can efficiently access essential information about 
a resource's current state, conditions, and other relevant details without fetching the entire resource, 
simplifying monitoring and aiding in prompt decision-making and issue resolution.

The new status subresource `latestTransaction` consists of a couple of fields that show the latest result of the 
reconciliation:
- `.status.latestTransaction.status`: shows the status of the latest reconciliation, either Successful or Failed;
- `.status.latestTransaction.time`: shows the time of the latest reconciliation;
- `.status.latestTransaction.error`: shows the message of an error if observed in the latest transaction.

## Example: Find out why an APIDefinition resource cannot be deleted
Consider the scenario when APIDefinition and SecurityPolicy are connected. Usually, APIDefinition cannot be deleted 
directly since it is protected by SecurityPolicy. The proper approach to remove an APIDefinition is to first remove 
the reference to the SecurityPolicy (either by deleting the SecurityPolicy CR or updating SecurityPolicy CR’s 
specification), and then remove the APIDefinition itself. However, if we directly delete this APIDefinition, Tyk 
Operator won’t delete the APIDefinition unless the link between SecurityPolicy and APIDefinition is removed. It is to 
protect the referential integrity between your resources.

```bash
$ kubectl delete tykapis httpbin 
apidefinition.tyk.tyk.io "httpbin" deleted 
^C%
```

After deleting APIDefinition, the operation hangs, and we suspect that something is wrong.
Users might still look through the logs to comprehend the issue, as they did in the past, but they can now examine 
their APIDefinition’s status subresource to make their initial, speedy issue diagnosis.

```bash
$ kubectl get tykapis httpbin 
NAME      DOMAIN   LISTENPATH   PROXY.TARGETURL      ENABLED   STATUS
httpbin            /httpbin     http://httpbin.org   true      Failed
```
As seen in the `STATUS` column, something went wrong, and the `STATUS` is `Failed`.

To get more information about the APIDefinition resource, we can use `kubectl describe` or `kubectl get`:

```bash
$ kubectl describe tykapis httpbin 
Name:         httpbin 
Namespace:    default 
API Version:  tyk.tyk.io/v1alpha1 
Kind:         ApiDefinition 
Metadata:
  ... 
Spec:
   ...
Status:
  api_id:                ZGVmYXVsdC9odHRwYmlu
  Latest CRD Spec Hash:  9169537376206027578
  Latest Transaction:
    Error:               unable to delete api due to security policy dependency=default/httpbin
    Status:              Failed
    Time:                2023-07-18T07:26:45Z
  Latest Tyk Spec Hash:  14558493065514264307
  linked_by_policies:
    Name:       httpbin
    Namespace:  default
```

or

```bash
$ kubectl get tykapis httpbin -o json | jq .status.latestTransaction
{
  "error": "unable to delete api due to security policy dependency=default/httpbin",
  "status": "Failed",
  "time": "2023-07-18T07:26:45Z"
}
```
Instead of digging into Tyk Operator's logs, we can now diagnose this issue simply by looking at the 
`.status.latestTransaction` field. As `.status.latestTransaction.error` implies, the error is related to SecurityPolicy 
dependency. 
