package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// TransactionStatus indicates the status of the Tyk API calls for currently reconciled object.
// Valid values are:
// - "Successful": showing that Tyk API calls on resource is completed successfully.
// - "Failed": showing that Tyk API calls on resource is failed.
// - "IngressTemplate": showing this resource is being used as template for Ingress controller.
type TransactionStatus string

const (
	// Successful shows that Tyk API calls on currently reconciled object finished successfully.
	Successful TransactionStatus = "Successful"

	// Failed shows that the operation on resource is failed due to Tyk API errors.
	Failed TransactionStatus = "Failed"

	// IngressTemplate shows that this resource is being used as template for Ingress controller.
	// Therefore, this resource is not going to be created at Tyk side. It will be used as a reference
	// for Ingress Controller.
	IngressTemplate TransactionStatus = "IngressTemplate"
)

// TransactionInfo holds information about the status of object's reconciliation.
type TransactionInfo struct {
	// Time corresponds to the time of last transaction.
	Time metav1.Time `json:"time,omitempty"`

	// Status corresponds to the status of the last transaction.
	Status TransactionStatus `json:"status,omitempty"`

	// Error corresponds to the error happened on Tyk API level, if any.
	Error string `json:"error,omitempty"`
}
