// Package config menyediakan konfigurasi untuk HTTP client yang digunakan
// untuk berkomunikasi dengan service eksternal. Package ini menggunakan
// pattern dependency injection dan functional options untuk fleksibilitas.
package config

import "github.com/parnurzeal/gorequest"

// ClientConfig adalah struktur yang menyimpan konfigurasi untuk HTTP client.
// Struktur ini menggunakan gorequest sebagai underlying HTTP client library.
type ClientConfig struct {
	// client adalah instance dari gorequest.SuperAgent yang digunakan untuk melakukan HTTP request
	client *gorequest.SuperAgent
	// baseURL adalah URL dasar untuk semua request yang akan dibuat
	baseURL string
	// signatureKey adalah kunci yang digunakan untuk autentikasi atau signing request
	signatureKey string
}

// IClientConfig adalah interface yang mendefinisikan kontrak untuk konfigurasi client.
// Interface ini memungkinkan dependency injection dan memudahkan testing dengan mock.
type IClientConfig interface {
	// Client mengembalikan instance gorequest.SuperAgent untuk melakukan HTTP request
	Client() *gorequest.SuperAgent
	// BaseURL mengembalikan URL dasar yang dikonfigurasi
	BaseURL() string
	// SignatureKey mengembalikan kunci signature yang dikonfigurasi
	SignatureKey() string
}

// Option adalah function type yang digunakan untuk mengkonfigurasi ClientConfig.
// Pattern ini memungkinkan konfigurasi yang fleksibel dan extensible.
type Option func(*ClientConfig)

// NewClientConfig membuat instance baru dari ClientConfig dengan konfigurasi default
// dan menerapkan options yang diberikan. Function ini menggunakan functional options pattern
// untuk memberikan fleksibilitas dalam konfigurasi.
//
// Parameters:
//   - options: variadic parameter berisi function-function untuk mengkonfigurasi client
//
// Returns:
//   - IClientConfig: interface yang menyediakan akses ke konfigurasi client
func NewClientConfig(options ...Option) IClientConfig {
	// Membuat instance ClientConfig dengan konfigurasi default
	clientConfig := &ClientConfig{
		// Inisialisasi gorequest client dengan header default untuk JSON communication
		client: gorequest.New().
			Set("Content_type", "application/json").
			Set("Accept", "application/json"),
	}

	// Menerapkan semua options yang diberikan untuk mengkustomisasi konfigurasi
	for _, option := range options {
		option(clientConfig)
	}

	return clientConfig
}

// Client mengembalikan instance gorequest.SuperAgent yang sudah dikonfigurasi.
// Method ini digunakan untuk mendapatkan HTTP client yang siap digunakan.
func (c *ClientConfig) Client() *gorequest.SuperAgent {
	return c.client
}

// BaseURL mengembalikan URL dasar yang dikonfigurasi untuk client.
// URL ini akan digunakan sebagai prefix untuk semua endpoint request.
func (c *ClientConfig) BaseURL() string {
	return c.baseURL
}

// SignatureKey mengembalikan kunci signature yang dikonfigurasi.
// Kunci ini digunakan untuk autentikasi atau signing HTTP request.
func (c *ClientConfig) SignatureKey() string {
	return c.signatureKey
}

// WithBaseURL adalah option function untuk mengatur base URL client.
// Function ini mengembalikan Option yang akan mengkonfigurasi baseURL.
//
// Parameters:
//   - baseURL: URL dasar yang akan digunakan untuk semua request
//
// Returns:
//   - Option: function yang akan mengaplikasikan konfigurasi baseURL
//
// Example:
//
//	config := NewClientConfig(WithBaseURL("https://api.example.com"))
func WithBaseURL(baseURL string) Option {
	return func(c *ClientConfig) {
		c.baseURL = baseURL
	}
}

// WithSignatureKey adalah option function untuk mengatur signature key client.
// Function ini mengembalikan Option yang akan mengkonfigurasi signatureKey.
//
// Parameters:
//   - signatureKey: kunci yang digunakan untuk autentikasi atau signing request
//
// Returns:
//   - Option: function yang akan mengaplikasikan konfigurasi signatureKey
//
// Example:
//
//	config := NewClientConfig(WithSignatureKey("your-secret-key"))
func WithSignatureKey(signatureKey string) Option {
	return func(c *ClientConfig) {
		c.signatureKey = signatureKey
	}
}
