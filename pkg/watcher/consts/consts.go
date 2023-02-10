package consts

const (
	DefaultLBPort          = ":443"
	AnnotationPrefix       = "synthetic-checker"
	FinalizerName          = AnnotationPrefix + "/finalizer"
	SkipAnnotation         = AnnotationPrefix + "/skip"
	IgnoreDeleteAnnotation = AnnotationPrefix + "/ignore-delete"
	TlsAnnotation          = AnnotationPrefix + "/TLS"
	NoTLSAnnotation        = AnnotationPrefix + "/noTLS"
	PortsAnnotation        = AnnotationPrefix + "/ports"
	IntervalAnnotation     = AnnotationPrefix + "/interval"
	EndpointsAnnotation    = AnnotationPrefix + "/endpoints"
	ConfigFromAnnotation   = AnnotationPrefix + "/configFrom"
)
