package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acmpca"
	"github.com/aws/aws-sdk-go-v2/service/acmpca/types"
)

var (
	currentCSR        []byte
	currentPrivateKey *rsa.PrivateKey
)

func main() {
	http.HandleFunc("/generate-csr", generateCSRHandler)
	http.HandleFunc("/issue-certificate", issueCertificateHandler)

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// generateCSRHandler generates a private key and CSR, and stores them in memory.
func generateCSRHandler(w http.ResponseWriter, r *http.Request) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		http.Error(w, "Failed to generate private key: "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentPrivateKey = privateKey

	subject := pkix.Name{
		CommonName:         "codecornersoftwares.co.za",
		Organization:       []string{"Code Corner"},
		OrganizationalUnit: []string{"Development"},
		Country:            []string{"ZA"},
		Province:           []string{"Cape Town"},
		Locality:           []string{"South Africa"},
	}
	csrTemplate := &x509.CertificateRequest{
		Subject:            subject,
		SignatureAlgorithm: x509.SHA256WithRSA,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
	if err != nil {
		http.Error(w, "Failed to generate CSR: "+err.Error(), http.StatusInternalServerError)
		return
	}

	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	currentCSR = csrPEM

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(csrPEM)
}

// issueCertificateHandler issues a certificate using the CSR and ACM PCA.
func issueCertificateHandler(w http.ResponseWriter, r *http.Request) {
	if len(currentCSR) == 0 {
		http.Error(w, "No CSR available. Please call /generate-csr first.", http.StatusBadRequest)
		return
	}

	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		http.Error(w, "Failed to load AWS config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.Region = "your-region"

	client := acmpca.NewFromConfig(cfg)

	input := &acmpca.IssueCertificateInput{
		CertificateAuthorityArn: aws.String("arn:aws:acm-pca:region:account-id:certificate-authority/CA-ID"),
		Csr:                     currentCSR,
		SigningAlgorithm:        types.SigningAlgorithmSha512withrsa,
		Validity: &types.Validity{
			Type:  types.ValidityPeriodTypeDays,
			Value: aws.Int64(365),
		},
	}

	resp, err := client.IssueCertificate(ctx, input)
	if err != nil {
		http.Error(w, "Error issuing certificate: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Certificate issued: %s", aws.ToString(resp.CertificateArn))
}
