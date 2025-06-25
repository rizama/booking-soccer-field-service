package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// Pagination
type PaginationParam struct {
	Count int64       `json:"count"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
	Data  interface{} `json:"data"`
}

type PaginationResult struct {
	TotalPage    int         `json:"totalPage"`
	TotalData    int         `json:"totalData"`
	NextPage     *int        `json:"nextPage"`
	PreviousPage *int        `json:"previousPage"`
	Page         int         `json:"page"`
	Limit        int         `json:"limit"`
	Data         interface{} `json:"data"`
}

func GeneratePagination(params PaginationParam) PaginationResult {
	totalPage := int(math.Ceil(float64(params.Count) / float64(params.Limit)))

	var (
		nextPage     int
		previousPage int
	)

	if params.Page < totalPage {
		nextPage = params.Page + 1
	}

	if params.Page > 1 {
		previousPage = params.Page - 1
	}

	result := PaginationResult{
		TotalPage:    totalPage,
		TotalData:    int(params.Count),
		NextPage:     &nextPage,
		PreviousPage: &previousPage,
		Page:         params.Page,
		Limit:        params.Limit,
		Data:         params.Data,
	}

	return result
}

// Generate SHA
func GenerateSHA256(inputString string) string {
	hash := sha256.New()
	hash.Write([]byte(inputString))
	hashByte := hash.Sum(nil)
	hashString := hex.EncodeToString(hashByte)
	return hashString
}

// Generate Rupiah Format
func GenerateRupiahFormat(amount *float64) string {
	stringValue := "0"
	if amount != nil {
		humanizeValue := humanize.CommafWithDigits(*amount, 0)
		stringValue = strings.ReplaceAll(humanizeValue, ",", ".")
	}

	return fmt.Sprintf("Rp. %s", stringValue)
}

// read file JSON
func BindFromJSON(destination any, filename, path string) error {
	viper := viper.New()

	viper.SetConfigType("json")
	viper.AddConfigPath(path)
	viper.SetConfigFile(filename)

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = viper.Unmarshal(&destination)
	if err != nil {
		logrus.Errorf("failed to unmarshal config file: %v", err)
		return err
	}

	return nil
}

func SetEnvFromConsulKV(v *viper.Viper) error {
	env := make(map[string]any)

	err := v.Unmarshal(&env)
	if err != nil {
		logrus.Errorf("failed to unmarshal config file: %v", err)
		return err
	}

	for k, v := range env {
		var (
			valOf = reflect.ValueOf(v)
			val   string
		)

		switch valOf.Kind() {
		case reflect.String:
			val = valOf.String()
		case reflect.Int:
			val = strconv.Itoa(int(valOf.Int()))
		case reflect.Uint:
			val = strconv.Itoa(int(valOf.Uint()))
		case reflect.Float32:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Float64:
			val = strconv.Itoa(int(valOf.Float()))
		case reflect.Bool:
			val = strconv.FormatBool(valOf.Bool())
		}

		err = os.Setenv(k, val)
		if err != nil {
			logrus.Errorf("failed to set env: %v", err)
			return err
		}
	}

	return nil

}

func BindFromConsul(destination any, endPoint, path string) error {
	viper := viper.New()

	viper.SetConfigType("json")
	err := viper.AddRemoteProvider("consul", endPoint, path)
	if err != nil {
		logrus.Errorf("failed to add remote provider: %v", err)
		return err
	}

	err = viper.ReadRemoteConfig()
	if err != nil {
		logrus.Errorf("failed to read remote config: %v", err)
		return err
	}

	err = viper.Unmarshal(&destination)
	if err != nil {
		logrus.Errorf("failed to unmarshal config file: %v", err)
		return err
	}

	err = SetEnvFromConsulKV(viper)
	if err != nil {
		logrus.Errorf("failed to set env from consul kv: %v", err)
		return err
	}

	return nil

}
