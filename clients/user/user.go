package clients

import (
	"context"
	clientConfig "field-service/clients/config"
	"field-service/common/utils"
	"field-service/config"
	"field-service/constants"
	"fmt"
	"net/http"
	"time"
)

/*
 * TUJUAN FILE INI:
 * File ini berfungsi sebagai HTTP client untuk berkomunikasi dengan User Service.
 * Digunakan untuk mengambil data user berdasarkan token autentikasi yang diberikan.
 *
 * KOMPONEN UTAMA:
 * 1. UserClient - struct yang mengimplementasikan komunikasi dengan User Service
 * 2. IUserClient - interface yang mendefinisikan contract untuk operasi user
 * 3. GetUserByToken - method untuk mendapatkan data user dari token
 *
 * FLOW AUTENTIKASI:
 * 1. Generate API key menggunakan SHA256 hash
 * 2. Set headers untuk autentikasi antar service
 * 3. Kirim request ke User Service
 * 4. Parse response dan return data user
 */

// UserClient struct yang menyimpan konfigurasi untuk komunikasi dengan User Service
type UserClient struct {
	client clientConfig.IClientConfig // HTTP client configuration untuk request
}

// IUserClient interface yang mendefinisikan contract untuk operasi user
type IUserClient interface {
	GetUserByToken(context.Context) (*UserData, error) // Method untuk mendapatkan data user dari token
}

// NewUserClient factory function untuk membuat instance UserClient baru
// Parameter: client - konfigurasi HTTP client yang akan digunakan
// Return: instance IUserClient yang siap digunakan
func NewUserClient(client clientConfig.IClientConfig) IUserClient {
	return &UserClient{
		client: client,
	}
}

// GetUserByToken method untuk mendapatkan data user dari User Service menggunakan token
// Method ini melakukan autentikasi antar service dan mengambil informasi user
// Parameter: ctx - context yang berisi token user untuk autentikasi
// Return: *UserData berisi informasi user, atau error jika gagal
func (u *UserClient) GetUserByToken(ctx context.Context) (*UserData, error) {
	// Step 1: Generate timestamp untuk API key security
	unixTime := time.Now().Unix()

	// Step 2: Buat string untuk generate API key dengan format: appName:signatureKey:timestamp
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		config.Config.AppName,   // Nama aplikasi dari config
		u.client.SignatureKey(), // Signature key untuk autentikasi antar service
		unixTime,                // Unix timestamp untuk mencegah replay attack
	)

	// Step 3: Generate SHA256 hash dari string API key untuk keamanan
	apiKey := utils.GenerateSHA256(generateAPIKey)

	// Step 4: Ambil token user dari context dan format sebagai Bearer token
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	// Step 5: Siapkan variable untuk menampung response dari User Service
	var response UserResponse

	// Step 6: Buat HTTP request dengan headers yang diperlukan untuk autentikasi antar service
	request := u.client.Client().Clone().
		Set(constants.Authorization, bearerToken).                  // Bearer token user
		Set(constants.XServiceName, config.Config.AppName).         // Nama service yang melakukan request
		Set(constants.XApiKey, apiKey).                             // API key untuk autentikasi antar service
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).     // Timestamp request
		Get(fmt.Sprintf("%s/api/v1/auth/user", u.client.BaseURL())) // Endpoint User Service

	// Step 7: Eksekusi request dan parse response ke struct UserResponse
	resp, _, errs := request.EndStruct(&response)
	if len(errs) > 0 {
		return nil, errs[0] // Return error jika ada masalah dalam request
	}

	// Step 8: Validasi status code response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user response: %s", response.Message)
	}

	// Step 9: Return data user jika berhasil
	return &response.Data, nil
}
