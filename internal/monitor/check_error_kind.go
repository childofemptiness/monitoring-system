package monitor

type CheckErrorKind string

const (
	CheckErrorKindUnknown CheckErrorKind = "unknown"
	CheckErrorTimeout     CheckErrorKind = "timeout"
	CheckErrorCanceled    CheckErrorKind = "canceled"
	CheckErrorConnection  CheckErrorKind = "connection"
	CheckErrorDNS         CheckErrorKind = "dns"
	CheckErrorTLS         CheckErrorKind = "tls"
)
