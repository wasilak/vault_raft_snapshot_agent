package config

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

// Configuration is the overall config object
type Configuration struct {
	Address         string      `mapstructure:"addr" json:"addr"`
	Retain          int64       `mapstructure:"retain" json:"retain"`
	Frequency       string      `mapstructure:"frequency" json:"frequency"`
	AWS             S3Config    `mapstructure:"aws_storage" json:"aws_storage"`
	Local           LocalConfig `mapstructure:"local_storage" json:"local_storage"`
	GCP             GCPConfig   `mapstructure:"google_storage" json:"google_storage"`
	Azure           AzureConfig `mapstructure:"azure_storage" json:"azure_storage"`
	RoleID          string      `mapstructure:"role_id" json:"role_id"`
	SecretID        string      `mapstructure:"secret_id" json:"secret_id"`
	Approle         string      `mapstructure:"approle" json:"approle"`
	K8sAuthRole     string      `mapstructure:"k8s_auth_role,omitempty" json:"k8s_auth_role,omitempty"`
	K8sAuthPath     string      `mapstructure:"k8s_auth_path,omitempty" json:"k8s_auth_path,omitempty"`
	VaultAuthMethod string      `mapstructure:"vault_auth_method,omitempty" json:"vault_auth_method,omitempty"`
	Daemon          bool        `mapstructure:"daemon" json:"daemon"`
}

// AzureConfig is the configuration for Azure blob snapshots
type AzureConfig struct {
	AccountName   string `mapstructure:"account_name" json:"account_name"`
	AccountKey    string `mapstructure:"account_key" json:"account_key"`
	ContainerName string `mapstructure:"container_name" json:"container_name"`
}

// GCPConfig is the configuration for GCP Storage snapshots
type GCPConfig struct {
	Bucket string `mapstructure:"bucket" json:"bucket"`
}

// LocalConfig is the configuration for local snapshots
type LocalConfig struct {
	Path string `mapstructure:"path" json:"path"`
}

// S3Config is the configuration for S3 snapshots
type S3Config struct {
	Uploader           *s3manager.Uploader
	AccessKeyID        string `mapstructure:"access_key_id" json:"access_key_id"`
	SecretAccessKey    string `mapstructure:"secret_access_key" json:"secret_access_key"`
	Endpoint           string `mapstructure:"s3_endpoint" json:"s3_endpoint"`
	Region             string `mapstructure:"s3_region" json:"s3_region"`
	Bucket             string `mapstructure:"s3_bucket" json:"s3_bucket"`
	KeyPrefix          string `mapstructure:"s3_key_prefix" json:"s3_key_prefix"`
	SSE                bool   `mapstructure:"s3_server_side_encryption" json:"s3_server_side_encryption"`
	StaticSnapshotName string `mapstructure:"s3_static_snapshot_name" json:"s3_static_snapshot_name"`
	S3ForcePathStyle   bool   `mapstructure:"s3_force_path_style" json:"s3_force_path_style"`
}

// ReadConfig reads the configuration file
func ReadConfig() (*Configuration, error) {
	viper.SetDefault("config", "./snapshot.json")
	viper.SetEnvPrefix("VRSA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	c := Configuration{}

	viper.SetConfigFile(viper.GetString("config"))
	viperErr := viper.ReadInConfig()
	if viperErr != nil { // Handle errors reading the config file
		slog.Error(viperErr.Error())
		panic(viperErr)
	}

	err := viper.Unmarshal(&c)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
		return nil, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return &c, nil
}
