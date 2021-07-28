package keys

// finalizers
const (
	ApiDefFinalizerName               = "finalizers.tyk.io/apidefinition"
	ApiDefTemplateFinalizerName       = "finalizers.tyk.io/template"
	PortalAPICatalogueFinalizerName   = "finalizers.tyk.io/portalapicatalogue"
	PortalAPIDescriptionFinalizerName = "finalizers.tyk.io/portalapidescription"
	PortalConfigurationFinalizerName  = "finalizers.tyk.io/portalconfiguration"
	GatewayAPIFinalizerName           = "finalizers.tyk.io/gatewayapi"
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

const GatewayAPI = "tyk.io/gateway-api"
