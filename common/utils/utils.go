package utils

import (
	"os"
	"reflect"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

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
