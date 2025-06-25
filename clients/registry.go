// Package clients menyediakan registry untuk mengelola semua HTTP client
// yang digunakan untuk berkomunikasi dengan service eksternal.
// Registry ini menggunakan pattern Factory dan Dependency Injection
// untuk memberikan akses terpusat ke berbagai client.
package clients

import (
	clientsConfig "field-service/clients/config"
	clientUser "field-service/clients/user"
	"field-service/config"
)

// ClientRegistry adalah struktur yang bertindak sebagai factory untuk semua HTTP client.
// Registry ini menyediakan akses terpusat ke berbagai client yang dikonfigurasi
// dengan pengaturan yang sesuai dari konfigurasi aplikasi.
type ClientRegistry struct{}

// IClientRegistry adalah interface yang mendefinisikan kontrak untuk registry client.
// Interface ini memungkinkan dependency injection dan memudahkan testing dengan mock.
// Setiap method mengembalikan client yang sudah dikonfigurasi dan siap digunakan.
type IClientRegistry interface {
	// UserSvc mengembalikan client untuk berkomunikasi dengan User Service
	UserSvc() clientUser.IUserClient
}

// NewClientRegistry membuat instance baru dari ClientRegistry.
// Function ini menggunakan pattern Factory untuk menyediakan akses
// ke registry yang sudah dikonfigurasi.
//
// Returns:
//   - IClientRegistry: interface yang menyediakan akses ke semua client
func NewClientRegistry() IClientRegistry {
	return &ClientRegistry{}
}

// UserSvc mengembalikan client yang dikonfigurasi untuk berkomunikasi dengan User Service.
// Method ini membuat instance baru dari UserClient dengan konfigurasi yang diambil
// dari konfigurasi aplikasi (config.Config).
//
// Returns:
//   - clientUser.IUserClient: client yang siap digunakan untuk User Service
//
// Configuration:
//   - BaseURL: diambil dari config.Config.InternalService.User.Host
//   - SignatureKey: diambil dari config.Config.InternalService.User.SignatureKey
func (c *ClientRegistry) UserSvc() clientUser.IUserClient {
	return clientUser.NewUserClient(
		clientsConfig.NewClientConfig(
			clientsConfig.WithBaseURL(config.Config.InternalService.User.Host),
			clientsConfig.WithSignatureKey(config.Config.InternalService.User.SignatureKey),
		),
	)
}
