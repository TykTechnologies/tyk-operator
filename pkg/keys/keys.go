package keys

// finalizers
const (
	EndpointsFinalizerName            = "finalizers.tyk.io/endpoints"
	ApiDefFinalizerName               = "finalizers.tyk.io/apidefinition"
	ApiDefTemplateFinalizerName       = "finalizers.tyk.io/template"
	PortalAPICatalogueFinalizerName   = "finalizers.tyk.io/portalapicatalogue"
	PortalAPIDescriptionFinalizerName = "finalizers.tyk.io/portalapidescription"
	PortalConfigurationFinalizerName  = "finalizers.tyk.io/portalconfiguration"
	OperatorContextFinalizerName      = "finalizers.tyk.io/operatorcontext"
	SubGraphFinalizerName             = "finalizers.tyk.io/subgraph"
	SuperGraphFinalizerName           = "finalizers.tyk.io/supergraph"
)

// Ingress
const (
	IngressLabel                       = "tyk.io/ingress"
	IngressTaintLabel                  = "tyk.io/taint"
	APIDefLabel                        = "tyk.io/apidefinition"
	IngressFinalizerName               = "finalizers.tyk.io/ingress"
	IngressClassAnnotation             = "kubernetes.io/ingress.class"
	IngressTemplateAnnotation          = "tyk.io/template"
	DefaultIngressClassAnnotationValue = "tyk"
)
