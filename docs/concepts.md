## Concepts

### Idempotency

Imagine any use case where you have keys tied to policies, and policies tied to APIs.

Now imagine that these resources are unintentionally destroyed.  Our database goes down, or our cluster, or something else.
 
Well, using the Tyk Operator, we can easily re-generate all our resources  in a non-destructive fashion.  That's because the operator
intelligently constructs the unique ID using the unique namespaced name of our CRD resources.  For that reason.

Alternatively, if you don't explicitly state it, it will be hard-coded for you by Base64 encoding the namespaced name of the CRD.

**For example**
1. we have keys tied to policies tied to APIs in production.
2. Our production DB gets destroyed, all our Policies and APIs are wiped
3. The Tyk Operator can resync all the changes from our CRDs into a new environment, by explicitly defining the Policy IDs and API IDs as before.
4. This allows keys to continue to work normally as Tyk resources are generated idempotently through the Operator.


### Migrating pre-Operator resources

Please visit the Readme of each individual CRD to read about how you can migrate an existing API or Policy resource into the Operator without any downtime or loss of functionality
 