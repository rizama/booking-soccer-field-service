package gcs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

/*
 * FUNGSI FILE INI:
 * File ini berfungsi sebagai wrapper untuk Google Cloud Storage (GCS) yang menyediakan
 * fungsionalitas upload file ke bucket GCS menggunakan service account authentication.
 * 
 * KOMPONEN UTAMA:
 * 1. ServiceAccountKeyJSON - struct untuk konfigurasi autentikasi GCS
 * 2. GCSClient - client untuk operasi GCS
 * 3. UploadFile - method untuk upload file ke GCS bucket
 */

// ServiceAccountKeyJSON represents the Google Cloud Service Account key configuration
// Struct ini menyimpan semua informasi yang diperlukan untuk autentikasi ke GCS
type ServiceAccountKeyJSON struct {
	Type                    string `json:"type"`                      // Tipe service account (biasanya "service_account")
	ProjectID               string `json:"project_id"`               // ID project Google Cloud
	PrivateKeyID            string `json:"private_key_id"`            // ID private key untuk autentikasi
	PrivateKey              string `json:"private_key"`               // Private key dalam format PEM
	ClientEmail             string `json:"client_email"`              // Email service account
	ClientID                string `json:"client_id"`                 // Client ID service account
	AuthURI                 string `json:"auth_uri"`                  // URI untuk autentikasi OAuth2
	TokenURI                string `json:"token_uri"`                 // URI untuk mendapatkan token
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"` // URL sertifikat X509 provider
	ClientX509CertURL       string `json:"client_x509_cert_url"`       // URL sertifikat X509 client
	UniverseDomain          string `json:"universe_domain"`            // Domain universe (biasanya "googleapis.com")
}

// GCSClient struct yang menyimpan konfigurasi untuk koneksi ke GCS
type GCSCLient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON // Konfigurasi autentikasi
	BucketName            string                // Nama bucket GCS target
}

// IGCSClient interface yang mendefinisikan contract untuk operasi GCS
type IGCSClient interface {
	UpdloadFile(context.Context, string, []byte) (string, error) // Method untuk upload file
}

// NewGCSClient factory function untuk membuat instance GCS client baru
// Parameter:
// - serviceAccountKeyJSON: konfigurasi autentikasi GCS
// - bucketName: nama bucket GCS yang akan digunakan
// Return: instance IGCSClient yang siap digunakan
func NewGCSClient(serviceAccountKeyJSON ServiceAccountKeyJSON, bucketName string) IGCSClient {
	return &GCSCLient{
		ServiceAccountKeyJSON: serviceAccountKeyJSON,
		BucketName:            bucketName,
	}
}

// createClient membuat dan menginisialisasi Google Cloud Storage client
// Method ini melakukan autentikasi menggunakan service account key JSON
// Parameter: ctx - context untuk operasi
// Return: *storage.Client yang sudah terotentikasi, atau error jika gagal
func (g *GCSCLient) createClient(ctx context.Context) (*storage.Client, error) {
	// Step 1: Membuat buffer untuk menyimpan JSON credentials
	reqBodyBytes := new(bytes.Buffer)

	// Step 2: Encode service account key JSON ke dalam buffer
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("Failed to encode service account key json: %v", err)
		return nil, err
	}

	// Step 3: Konversi buffer ke byte array untuk credentials
	jsonByte := reqBodyBytes.Bytes()
	
	// Step 4: Membuat GCS client dengan credentials JSON
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonByte))
	if err != nil {
		logrus.Errorf("Failed to create client: %v", err)
		return nil, err
	}

	// Step 5: Return client yang sudah siap digunakan
	return client, nil
}

// UpdloadFile method untuk upload file ke Google Cloud Storage bucket
// Method ini melakukan upload file dengan langkah-langkah yang aman dan terstruktur
// Parameter:
// - ctx: context untuk operasi (bisa dibatalkan)
// - fileName: nama file yang akan disimpan di bucket
// - data: byte array dari file yang akan diupload
// Return: URL publik file yang berhasil diupload, atau error jika gagal
func (c *GCSCLient) UpdloadFile(ctx context.Context, fileName string, data []byte) (string, error) {
	// Step 1: Konfigurasi default untuk upload
	var (
		contentType      = "application/octet-stream" // Content type default untuk file binary
		timeoutInSeconds = 60                         // Timeout 60 detik untuk operasi upload
	)

	// Step 2: Membuat GCS client dengan autentikasi
	client, err := c.createClient(ctx)
	if err != nil {
		logrus.Errorf("Failed to create client: %v", err)
		return "", err
	}

	// Step 3: Pastikan client ditutup setelah operasi selesai (resource cleanup)
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("Failed to close client: %v", err)
			return
		}
	}(client)

	// Step 4: Set timeout untuk operasi upload (mencegah hanging)
	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	// Step 5: Mendapatkan referensi ke bucket dan object yang akan dibuat
	bucket := client.Bucket(c.BucketName)  // Referensi ke bucket GCS
	object := bucket.Object(fileName)      // Referensi ke object/file dalam bucket
	buffer := bytes.NewBuffer(data)        // Buffer untuk data file

	// Step 6: Membuat writer untuk upload file ke GCS
	writer := object.NewWriter(ctx)
	writer.ChunkSize = 0 // Set chunk size 0 untuk upload dalam satu chunk

	// Step 7: Copy data dari buffer ke GCS object writer
	_, err = io.Copy(writer, buffer)
	if err != nil {
		logrus.Errorf("failed to copy: %v", err)
		return "", err
	}

	// Step 8: Tutup writer untuk finalisasi upload
	err = writer.Close()
	if err != nil {
		logrus.Errorf("failed to close: %v", err)
		return "", err
	}

	// Step 9: Update metadata object dengan content type yang sesuai
	_, err = object.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("failed to update: %v", err)
		return "", err
	}

	// Step 10: Generate URL publik untuk mengakses file yang sudah diupload
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", c.BucketName, fileName)
	return url, nil
}
