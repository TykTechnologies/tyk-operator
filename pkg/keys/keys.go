package keys

// APIDefinition
const (
	ApiDefFinalizerName         = "finalizers.tyk.io/apidefinition"
	ApiDefTemplateFinalizerName = "finalizers.tyk.io/template"
)

//Ingress
const (
	IngressLabelKey                    = "tyk.io/ingress"
	IngressTaintLabelKey               = "tyk.io/taint"
	APIDefLabelKey                     = "tyk.io/apidefinition"
	IngressFinalizerName               = "finalizers.tyk.io/ingress"
	IngressClassAnnotationKey          = "kubernetes.io/ingress.class"
	IngressTemplateAnnotationKey       = "tyk.io/template"
	DefaultIngressClassAnnotationValue = "tyk"
)
