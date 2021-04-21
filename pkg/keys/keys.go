package keys

// APIDefinition
const (
	ApiDefFinalizerName         = "finalizers.tyk.io/apidefinition"
	ApiDefTemplateFinalizerName = "finalizers.tyk.io/template"
	PortalAPIFinalizerName      = "finalizers.tyk.io/portalapi"
)

//Ingress
const (
	IngressLabel                       = "tyk.io/ingress"
	IngressTaintLabel                  = "tyk.io/taint"
	APIDefLabel                        = "tyk.io/apidefinition"
	IngressFinalizerName               = "finalizers.tyk.io/ingress"
	IngressClassAnnotation             = "kubernetes.io/ingress.class"
	IngressTemplateAnnotation          = "tyk.io/template"
	DefaultIngressClassAnnotationValue = "tyk"
)
