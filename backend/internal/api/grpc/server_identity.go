package grpc

import (
	"context"
	"crypto/x509"
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// ==================== Identity Extraction ====================

// MetadataKey constants for gRPC metadata.
const (
	// MetadataKeyClientCertDN is the client certificate full Subject DN.
	// Used for backward compatibility when behind TLS-terminating proxy.
	MetadataKeyClientCertDN = "x-client-cert-dn"

	// MetadataKeyClientCertSerial is the client certificate serial number.
	MetadataKeyClientCertSerial = "x-client-cert-serial"

	// MetadataKeyClientCertFingerprint is the client certificate fingerprint.
	MetadataKeyClientCertFingerprint = "x-client-cert-fingerprint"

	// MetadataKeyOrgSlug is the organization slug sent by Runner.
	MetadataKeyOrgSlug = "x-org-slug"

	// MetadataKeyRealIP is the real client IP.
	MetadataKeyRealIP = "x-real-ip"
)

// ClientIdentity holds information extracted from TLS peer or gRPC metadata.
type ClientIdentity struct {
	NodeID           string // From certificate CN
	OrgSlug          string // From certificate O or Runner metadata
	CertSerialNumber string // From certificate
	CertFingerprint  string // From certificate
	RealIP           string // Client IP
}

// ExtractClientIdentity extracts client identity from gRPC context.
// It first tries to extract from TLS peer certificate (direct mTLS mode),
// then falls back to metadata (proxy mode for backward compatibility).
func ExtractClientIdentity(ctx context.Context) (*ClientIdentity, error) {
	// Try to extract from TLS peer certificate first (direct mTLS mode)
	if identity, err := extractFromTLSPeer(ctx); err == nil {
		return identity, nil
	}

	// Fall back to metadata extraction (proxy mode)
	return extractFromMetadata(ctx)
}

// extractFromTLSPeer extracts client identity from TLS peer certificate.
// This is used when the server handles mTLS directly (TLS passthrough mode).
func extractFromTLSPeer(ctx context.Context) (*ClientIdentity, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no peer in context")
	}

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, fmt.Errorf("no TLS info in peer")
	}

	if len(tlsInfo.State.VerifiedChains) == 0 || len(tlsInfo.State.VerifiedChains[0]) == 0 {
		return nil, fmt.Errorf("no verified client certificate")
	}

	clientCert := tlsInfo.State.VerifiedChains[0][0]
	return extractFromCertificate(ctx, clientCert)
}

// extractFromCertificate extracts client identity from an X.509 certificate.
func extractFromCertificate(ctx context.Context, cert *x509.Certificate) (*ClientIdentity, error) {
	identity := &ClientIdentity{
		NodeID:           cert.Subject.CommonName,
		CertSerialNumber: cert.SerialNumber.String(),
	}

	// Extract organization from certificate (first one)
	if len(cert.Subject.Organization) > 0 {
		identity.OrgSlug = cert.Subject.Organization[0]
	}

	// Get org slug from metadata if not in certificate
	if identity.OrgSlug == "" {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get(MetadataKeyOrgSlug); len(values) > 0 {
				identity.OrgSlug = values[0]
			}
		}
	}

	// Validate required fields
	if identity.NodeID == "" {
		return nil, fmt.Errorf("missing client certificate CN (node_id)")
	}
	if identity.OrgSlug == "" {
		return nil, fmt.Errorf("missing org slug")
	}

	// Extract peer address if available
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		identity.RealIP = p.Addr.String()
	}

	return identity, nil
}

// extractFromMetadata extracts client identity from gRPC metadata.
// This is used for backward compatibility with TLS-terminating proxies.
func extractFromMetadata(ctx context.Context) (*ClientIdentity, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata in context")
	}

	identity := &ClientIdentity{}

	// Extract certificate DN and parse CN (node_id) - required
	if values := md.Get(MetadataKeyClientCertDN); len(values) > 0 && values[0] != "" {
		identity.NodeID = extractCNFromDN(values[0])
	}
	if identity.NodeID == "" {
		return nil, fmt.Errorf("missing client certificate CN (node_id)")
	}

	// Extract org slug - required
	if values := md.Get(MetadataKeyOrgSlug); len(values) > 0 {
		identity.OrgSlug = values[0]
	}
	if identity.OrgSlug == "" {
		return nil, fmt.Errorf("missing org slug")
	}

	// Extract optional fields
	if values := md.Get(MetadataKeyClientCertSerial); len(values) > 0 {
		identity.CertSerialNumber = values[0]
	}
	if values := md.Get(MetadataKeyClientCertFingerprint); len(values) > 0 {
		identity.CertFingerprint = values[0]
	}
	if values := md.Get(MetadataKeyRealIP); len(values) > 0 {
		identity.RealIP = values[0]
	}

	return identity, nil
}

// extractCNFromDN extracts Common Name (CN) from X.509 Subject DN string.
// Supports both formats:
// - OpenSSL default: "/CN=dev-runner/O=AgentsMesh/OU=Runner"
// - RFC 2253: "CN=dev-runner,O=AgentsMesh,OU=Runner"
func extractCNFromDN(dn string) string {
	if dn == "" {
		return ""
	}

	// Try RFC 2253 format first (comma-separated)
	for _, part := range splitDN(dn) {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToUpper(part), "CN=") {
			return strings.TrimPrefix(part, part[:3]) // Handle case variations
		}
	}

	return ""
}

// splitDN splits a DN string by comma or slash separators.
func splitDN(dn string) []string {
	// Check which format is used
	if strings.Contains(dn, "/") && !strings.Contains(dn, ",") {
		// OpenSSL format: "/CN=value/O=value"
		parts := strings.Split(dn, "/")
		// Filter out empty parts
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	}
	// RFC 2253 format: "CN=value,O=value"
	return strings.Split(dn, ",")
}
