package signerxml

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"

	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cgisoftware/initializers/signerxml/types"
	"software.sslmate.com/src/go-pkcs12"
)

type signerXml struct{}

// NewSignerXml creates a new instance of the XML signerco
func NewSignerXml() signerXml {
	return signerXml{}
}

// GetCertificateInfo extracts information from the PFX certificate
func (a signerXml) GetCertificateInfo(cert types.A1) (types.A1Info, error) {
	_, certificate, err := a.extractPfxCertificate(cert)
	if err != nil {
		return types.A1Info{}, err
	}

	info := types.A1Info{
		Subject:      certificate.Subject.String(),
		Issuer:       certificate.Issuer.String(),
		SerialNumber: certificate.SerialNumber.String(),
		NotBefore:    certificate.NotBefore,
		NotAfter:     certificate.NotAfter,
		Certificate:  certificate,
	}

	return info, nil
}

// SignXML signs an XML document using an A1 PFX certificate following the XMLDSIG standard
func (a signerXml) SignXML(dados types.Signature) (types.SignatureResult, error) {
	dados.XMLContent = cleanXML(dados.XMLContent)

	resultado := types.SignatureResult{
		Date:       time.Now(),
		Success:    false,
		XMLContent: dados.XMLContent,
	}

	// Validate input
	if err := a.validateInput(dados); err != nil {
		resultado.Error = err
		return resultado, err
	}

	// Extract certificate and private key from PFX
	privateKey, cert, err := a.extractPfxCertificate(dados.Certificate)
	if err != nil {
		resultado.Error = fmt.Errorf("Error extracting PFX certificate: %v", err)
		return resultado, err
	}

	// Validate certificate
	if err := a.validateCertificate(cert); err != nil {
		resultado.Error = fmt.Errorf("Invalid certificate: %v", err)
		return resultado, err

	}

	// Prepare XML for signing
	xmlPreparado, elementoID, err := a.prepareXMLForSigning(dados.XMLContent, dados.ElementID)
	if err != nil {
		resultado.Error = fmt.Errorf("Error preparing XML: %v", err)
		return resultado, err
	}

	// Calculate hash of the element to be signed
	hash, err := a.calculateElementHash(xmlPreparado, elementoID)
	if err != nil {
		resultado.Error = fmt.Errorf("Error calculating hash: %v", err)
		return resultado, err
	}

	// Create SignedInfo
	signedInfo, err := a.createSignedInfo(hash, elementoID)
	if err != nil {
		resultado.Error = fmt.Errorf("Error creating SignedInfo: %v", err)
		return resultado, err
	}

	// Canonicalize SignedInfo before signing
	// IMPORTANT: XMLDSIG validation checks the signature over the SignedInfo in its canonical form
	signedInfo = a.canonicalizeXML(signedInfo)

	// Sign SignedInfo
	assinatura, err := a.signSignedInfo(signedInfo, privateKey)
	if err != nil {
		resultado.Error = fmt.Errorf("Error signing: %v", err)
		return resultado, err
	}

	// Create complete Signature element
	signature, err := a.createElementSignature(signedInfo, assinatura, cert)
	if err != nil {
		resultado.Error = fmt.Errorf("Error creating Signature element: %v", err)
		return resultado, err
	}

	// Insert signature into XML
	xmlAssinado, err := a.insertSignatureInXML(xmlPreparado, signature, dados.TagSignature)
	if err != nil {
		resultado.Error = fmt.Errorf("Error inserting signature: %v", err)
		return resultado, err
	}

	resultado.XMLSigned = xmlAssinado
	resultado.Success = true
	return resultado, nil
}

// validateInput validates the input data
func (a signerXml) validateInput(dados types.Signature) error {
	if strings.TrimSpace(dados.XMLContent) == "" {
		return errors.New("XML content cannot be empty")
	}
	if len(dados.Certificate.File) == 0 {
		return errors.New("certificate file cannot be empty")
	}
	if strings.TrimSpace(dados.Certificate.Password) == "" {
		return errors.New("certificate password cannot be empty")
	}
	return nil
}

func (a signerXml) extractPfxCertificate(cert types.A1) (crypto.PrivateKey, *x509.Certificate, error) {
	blocks, err := pkcs12.ToPEM(cert.File, cert.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("error decoding PFX (ToPEM): %w", err)
	}

	var privateKey crypto.PrivateKey
	var certificate *x509.Certificate

	for _, block := range blocks {
		if block.Type == "CERTIFICATE" {
			c, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}
			// Assume the first certificate found or improve logic to find the leaf
			if certificate == nil {
				certificate = c
			}
		} else if block.Type == "PRIVATE KEY" || block.Type == "RSA PRIVATE KEY" {
			if k, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				privateKey = k
			} else if k, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
				privateKey = k
			}
		}
	}

	if privateKey == nil {
		return nil, nil, errors.New("private key not found in PFX file")
	}

	if certificate == nil {
		return nil, nil, errors.New("certificate not found in PFX file")
	}

	return privateKey, certificate, nil
}

// validateCertificate validates if the certificate is valid
func (a signerXml) validateCertificate(cert *x509.Certificate) error {
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		return errors.New("certificate expired")
	}
	return nil
}

// prepareXMLForSigning prepares the XML and identifies the element to be signed
func (a signerXml) prepareXMLForSigning(xmlContent, elementoID string) (string, string, error) {
	// If no ID was specified, look for the first element with an ID
	if elementoID == "" {
		re := regexp.MustCompile(`(?i)id\s*=\s*["']([^"']+)["']`)
		matches := re.FindStringSubmatch(xmlContent)
		if len(matches) > 1 {
			elementoID = matches[1]
		} else {
			return "", "", errors.New("element ID not found in XML")
		}
	}

	return xmlContent, elementoID, nil
}

// calculateElementHash calculates the SHA-256 hash of the element to be signed
func (a signerXml) calculateElementHash(xmlContent, elementoID string) (string, error) {
	// Extract the specific element by ID
	elemento, err := a.extractElementByID(xmlContent, elementoID)
	if err != nil {
		return "", fmt.Errorf("error extracting element %s: %v", elementoID, err)
	}

	// Apply transformations (enveloped-signature and canonicalization)
	elementoTransformado := a.applyTransformations(elemento)

	// Calculate SHA256 hash
	hasher := sha256.New()
	hasher.Write([]byte(elementoTransformado))
	hash := hasher.Sum(nil)
	return base64.StdEncoding.EncodeToString(hash), nil
}

// extractElementByID extracts a specific element from the XML by its ID
func (a signerXml) extractElementByID(xmlContent, elementoID string) (string, error) {
	// 1. Find the opening tag containing the ID
	// Capture the tag name (group 1)
	patternOpen := fmt.Sprintf(`<([\w:-]+)[^>]*Id=["']?%s["']?[^>]*>`, regexp.QuoteMeta(elementoID))
	reOpen := regexp.MustCompile(patternOpen)

	loc := reOpen.FindStringIndex(xmlContent)
	if loc == nil {
		// Try with lowercase id if failed
		patternOpen = fmt.Sprintf(`<([\w:-]+)[^>]*id=["']?%s["']?[^>]*>`, regexp.QuoteMeta(elementoID))
		reOpen = regexp.MustCompile(patternOpen)
		loc = reOpen.FindStringIndex(xmlContent)
	}

	if loc == nil {
		return "", fmt.Errorf("element with ID '%s' not found", elementoID)
	}

	// Extract the tag name
	matchOpen := xmlContent[loc[0]:loc[1]]
	submatches := reOpen.FindStringSubmatch(matchOpen)
	if len(submatches) < 2 {
		return "", fmt.Errorf("unable to identify tag name for ID '%s'", elementoID)
	}
	tagName := submatches[1]

	// 2. Extract the complete content up to the corresponding closing tag
	// Assuming no nested tags with the same name (common in NFe/NFSe XML for main blocks)
	patternFull := fmt.Sprintf(`(?s)%s.*?</%s>`, regexp.QuoteMeta(matchOpen), regexp.QuoteMeta(tagName))
	reFull := regexp.MustCompile(patternFull)

	matches := reFull.FindString(xmlContent)
	if matches == "" {
		return "", fmt.Errorf("unable to extract complete content of element '%s'", tagName)
	}

	if !strings.Contains(matches, "xmlns=") {
		reNS := regexp.MustCompile(`xmlns=["']([^"']+)["']`)
		matchNS := reNS.FindStringSubmatch(xmlContent)
		if len(matchNS) > 1 {
			namespace := matchNS[1]
			matches = strings.Replace(matches, tagName, tagName+fmt.Sprintf(" xmlns=\"%s\"", namespace), 1)
		}
	}

	return matches, nil
}

// applyTransformations applies necessary transformations to the element
func (a signerXml) applyTransformations(elemento string) string {
	// 1. Enveloped-signature transformation: remove existing signatures
	elementoSemAssinatura := a.removeSignatures(elemento)

	// 2. C14N Canonicalization
	elementoCanonical := a.canonicalizeXML(elementoSemAssinatura)

	return elementoCanonical
}

// removeSignatures removes signature elements from the XML
func (a signerXml) removeSignatures(xmlContent string) string {
	// Remove complete Signature elements
	re := regexp.MustCompile(`(?s)<Signature[^>]*>.*?</Signature>`)
	return re.ReplaceAllString(xmlContent, "")
}

// createSignedInfo creates the XMLDSIG SignedInfo element
func (a signerXml) createSignedInfo(digestValue, referenceURI string) (string, error) {
	signedInfo := fmt.Sprintf(`<SignedInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
		<CanonicalizationMethod Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></CanonicalizationMethod>
		<SignatureMethod Algorithm="http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"></SignatureMethod>
		<Reference URI="#%s">
			<Transforms>
				<Transform Algorithm="http://www.w3.org/2000/09/xmldsig#enveloped-signature"></Transform>
				<Transform Algorithm="http://www.w3.org/TR/2001/REC-xml-c14n-20010315"></Transform>
			</Transforms>
			<DigestMethod Algorithm="http://www.w3.org/2001/04/xmlenc#sha256"></DigestMethod>
			<DigestValue>%s</DigestValue>
		</Reference>
	</SignedInfo>`, referenceURI, digestValue)

	return signedInfo, nil
}

// signSignedInfo signs the SignedInfo element
func (a signerXml) signSignedInfo(signedInfo string, privateKey crypto.PrivateKey) (string, error) {
	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("private key must be RSA")
	}

	hasher := sha256.New()
	hasher.Write([]byte(signedInfo))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("error signing: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// createElementSignature creates the complete Signature element
func (a signerXml) createElementSignature(signedInfo, signatureValue string, cert *x509.Certificate) (string, error) {
	// Encode certificate in Base64
	certBase64 := base64.StdEncoding.EncodeToString(cert.Raw)

	signature := fmt.Sprintf(`<Signature xmlns="http://www.w3.org/2000/09/xmldsig#">
		%s
		<SignatureValue>%s</SignatureValue>
		<KeyInfo>
			<X509Data>
				<X509Certificate>%s</X509Certificate>
			</X509Data>
		</KeyInfo>
	</Signature>`, signedInfo, signatureValue, certBase64)

	return signature, nil
}

// insertSignatureInXML inserts the signature into the XML
func (a signerXml) insertSignatureInXML(xmlContent, signature, tagAssinatura string) (string, error) {
	// If no tag was specified, insert before the closing of the root element
	if tagAssinatura == "" {
		// Find the last closing tag
		re := regexp.MustCompile(`</([^>]+)>\s*$`)
		matches := re.FindStringSubmatch(xmlContent)
		if len(matches) > 1 {
			tagAssinatura = matches[1]
		} else {
			return "", errors.New("unable to determine where to insert the signature")
		}
	}

	// Insert signature before the specified closing tag
	closeTag := fmt.Sprintf("</%s>", tagAssinatura)
	xmlAssinado := strings.Replace(xmlContent, closeTag, signature+"\n"+closeTag, 1)

	return xmlAssinado, nil
}

// canonicalizeXML applies C14N canonicalization to the XML
func (a signerXml) canonicalizeXML(xmlContent string) string {
	canonical := regexp.MustCompile(`>\s+<`).ReplaceAllString(xmlContent, "><")
	// canonical = strings.TrimSpace(canonical)
	// canonical = regexp.MustCompile(`\s+`).ReplaceAllString(canonical, " ")

	return canonical
}

// readPFXCertificate reads a PFX certificate from a specified directory
func (a signerXml) ReadPFXCertificate(path string, senha string) (types.A1, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return types.A1{}, fmt.Errorf("directory not found: %s", path)
	}

	certificadoBytes, err := os.ReadFile(path)
	if err != nil {
		return types.A1{}, fmt.Errorf("error loading PFX certificate %s: %v", path, err)
	}

	cert := types.A1{
		File:     certificadoBytes,
		Password: senha,
	}

	_, cert1, err := a.extractPfxCertificate(cert)
	if err != nil {
		return types.A1{}, fmt.Errorf("error extracting PFX certificate: %v", err)
	}

	err = a.validateCertificate(cert1)
	if err != nil {
		return types.A1{}, fmt.Errorf("error validating certificate: %v", err)
	}

	return cert, nil
}

// readPFXCertificate reads a PFX certificate from a specified directory
func (a signerXml) ReadPFXCertificateFromBytes(certificadoBytes []byte, senha string) (types.A1, error) {
	cert := types.A1{
		File:     certificadoBytes,
		Password: senha,
	}

	_, cert1, err := a.extractPfxCertificate(cert)
	if err != nil {
		return types.A1{}, fmt.Errorf("error extracting PFX certificate: %v", err)
	}

	err = a.validateCertificate(cert1)
	if err != nil {
		return types.A1{}, fmt.Errorf("error validating certificate: %v", err)
	}

	return cert, nil
}

func cleanXML(xmlContent string) string {
	re := regexp.MustCompile(`>\s+<`)
	xmlContent = re.ReplaceAllString(xmlContent, "><")

	lines := strings.Split(xmlContent, "\n")
	var sb strings.Builder
	for _, line := range lines {
		line = strings.TrimSpace(line)
		sb.WriteString(line)
	}
	xmlContent = sb.String()

	removerChars := func(r rune) rune {
		if r == '\t' || r == '\r' {
			return -1
		}
		return r
	}
	xmlContent = strings.Map(removerChars, xmlContent)
	return xmlContent
}
