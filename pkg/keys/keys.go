package keys

// finalizers
const (
	ApiDefFinalizerName               = "finalizers.tyk.io/apidefinition"
	ApiDefTemplateFinalizerName       = "finalizers.tyk.io/template"
	PortalAPICatalogueFinalizerName   = "finalizers.tyk.io/portalapicatalogue"
	PortalAPIDescriptionFinalizerName = "finalizers.tyk.io/portalapidescription"
	PortalConfigurationFinalizerName  = "finalizers.tyk.io/portalconfiguration"
	OperatorContextFinalizerName      = "finalizers.tyk.io/operatorcontext"
	SubGraphFinalizerName             = "finalizers.tyk.io/subgraph"
	SuperGraphFinalizerName           = "finalizers.tyk.io/supergraph"
	TykOASFinalizerName               = "finalizers.tyk.io/tykoas"
	IngressFinalizerName              = "finalizers.tyk.io/ingress"
)

// Ingress
const (
	IngressClassAnnotation             = "kubernetes.io/ingress.class"
	IngressLabel                       = "tyk.io/ingress"
	IngressTaintLabel                  = "tyk.io/taint"
	APIDefLabel                        = "tyk.io/apidefinition"
	TykOASApiDefinitionLabel           = "tyk.io/tykoasapidefinition"
	IngressTemplateAnnotation          = "tyk.io/template"
	IngressTemplateKindAnnotation      = "tyk.io/template-kind"
	DefaultIngressClassAnnotationValue = "tyk"
	TykOasApiDefinitionTemplateLabel   = "tyk.io/template"
)
