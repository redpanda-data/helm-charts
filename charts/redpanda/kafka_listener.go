package redpanda

// Commenting conventions:
// {GoFieldName} User facing documentation
// ---
// Developer facing documentation
// Annotations +kubebuilder,+gotohelm,etc

// JSONScalar is any scalar JSON value marshalled to a string.
// +kubebuilder:validation:Pattern=^(\d+)|(true|false)|(".*")$
type JSONScalar string

type BrokerConfig struct {
	// KafkaListeners is an amalgam of `kafka_api`, `kafka_api_tls`,
	// `advertised_kafka_api`. That may be configured in a more ergonomic and
	// less brittle fashion. `{,advertised_}kafka_api{,_tls}` may be set
	// directly but are mutually exclusive to KafkaListeners.
	KafkaListeners []KafkaListener `json:"kafka_listeners,omitempty"`
	// KafkaAPI is a direct mapping to `kafka_api`. It takes precedence over
	// everything else, if provided. It is recommend to instead use .KafkaListeners.
	// +verbatim
	KafkaAPI []map[string]JSONScalar `json:"kafka_api,omitempty"`
	// AdvertisedKafkaAPI is a direct mapping to `advertised_kafka_api`. It
	// takes precedence over everything else, if provided. It is recommend to
	// instead use .KafkaListeners.
	// +verbatim
	AdvertisedKafkaAPI []map[string]JSONScalar `json:"advertised_kafka_api,omitempty"`
	// KafkaAPITLS is a direct mapping to `kafka_api_tls`. It takes precedence
	// over everything else, if provided. It is recommend to instead use
	// .KafkaListeners.
	// +verbatim
	KafkaAPITLS []map[string]JSONScalar `json:"kafka_api_tls,omitempty"`
}

// ---
// Enum definition: https://github.com/redpanda-data/redpanda/blob/8d5d1cc6d56d77b575f69150140aa5689bb75c47/src/v/config/broker_authn_endpoint.h#L30
// +kubebuilder:validation:Enum=none;sasl;mtls_identity
type KafkaAuthenticationMethod string

// See also: https://docs.redpanda.com/current/manage/security/listener-configuration/
type KafkaListener struct {
	// Name maps to the `kafka_api[*].name` field.
	// Must be unique across all kafka listeners.
	// +kubebuilder:validation:MaxLength=15
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
	// Port maps to the `kafka_api[*].port` field.
	// Must be unique across all listeners.
	// +kubebuilder:validation:Minimum=1
	Port int32 `json:"port"`
	// Address maps to the `kafka_api[*].address` field.
	// Generally doesn't need to be set.
	// +kubebuilder:default="0.0.0.0"
	Address *string `json:"address"`
	// AuthenticationMethod maps to the `kafka_api[*].authentication_method`
	// field.
	// +kubebuilder:default=none
	AuthenticationMethod KafkaAuthenticationMethod
	// TLS, if set, configures the corresponding `kafka_api_tls` element for this
	// `kafka_api` element.
	// TLS *KafkaListenerTLS
	TLS *KafkaListenerTLS
	// AdvertisedAddress, if set, maps to the `advertised_kafka_api[*].address`
	// field.
	AdvertisedAddress *PerBrokerValue[string]
	// AdvertisedPort, if set, maps to the `advertised_kafka_api[*].port`
	// field.
	AdvertisedPort *PerBrokerValue[int32]
}

type KafkaListenerTLS struct {
	Config     *KafkaListenerTLSConfig
	TLSCertRef *TLSCertReference
	// TODO Nice to have, just provide a reference to a cert-manager reference
	// and we'll configure it for you.
	// CertificateRef *CertificateReference
}

// See also https://docs.redpanda.com/current/manage/security/encryption/#configure-tls
type KafkaListenerTLSConfig struct {
	// Enabled controls the `kafka_api_tls[*].enabled` field. It is passed
	// through verbatim and does not affect any other configuration values.
	// +kubebuilder:default=true
	// +verbatim
	Enabled *bool

	// KeyFile maps to the `kafka_api_tls[*].key_file` field.
	KeyFile FileSource

	// CertFile maps to the `kafka_api_tls[*].cert_file` field.
	CertFile FileSource

	// TrustStoreFile maps to the `kafka_api_tls[*].truststore_file` field.
	// +kubebuilder:default={"path": "/etc/ssl/certs/ca-certificates.crt"}
	TrustStoreFile *FileSource

	// RequireClientAuth maps to the `kafka_api_tls[*].required_client_auth` field.
	RequireClientAuth bool `json:"required_client_auth"`
}
