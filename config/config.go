package config

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configuration is the overall config object
type Configuration struct {
	Address         string      `json:"addr"`
	Retain          int64       `json:"retain"`
	Frequency       string      `json:"frequency"`
	AWS             S3Config    `json:"aws_storage"`
	Local           LocalConfig `json:"local_storage"`
	GCP             GCPConfig   `json:"google_storage"`
	Azure           AzureConfig `json:"azure_storage"`
	RoleID          string      `json:"role_id"`
	SecretID        string      `json:"secret_id"`
	Approle         string      `json:"approle"`
	K8sAuthRole     string      `json:"k8s_auth_role,omitempty"`
	K8sAuthPath     string      `json:"k8s_auth_path,omitempty"`
	VaultAuthMethod string      `json:"vault_auth_method,omitempty"`
	Daemon          bool        `json:"daemon"`
}

// AzureConfig is the configuration for Azure blob snapshots
type AzureConfig struct {
	AccountName   string `json:"account_name"`
	AccountKey    string `json:"account_key"`
	ContainerName string `json:"container_name"`
}

// GCPConfig is the configuration for GCP Storage snapshots
type GCPConfig struct {
	Bucket string `json:"bucket"`
}

// LocalConfig is the configuration for local snapshots
type LocalConfig struct {
	Path string `json:"path"`
}

// S3Config is the configuration for S3 snapshots
type S3Config struct {
	Uploader           *s3manager.Uploader
	AccessKeyID        string `json:"access_key_id"`
	SecretAccessKey    string `json:"secret_access_key"`
	Endpoint           string `json:"s3_endpoint"`
	Region             string `json:"s3_region"`
	Bucket             string `json:"s3_bucket"`
	KeyPrefix          string `json:"s3_key_prefix"`
	SSE                bool   `json:"s3_server_side_encryption"`
	StaticSnapshotName string `json:"s3_static_snapshot_name"`
	S3ForcePathStyle   bool   `json:"s3_force_path_style"`
}

// ReadConfig reads the configuration file
func ReadConfig() (*Configuration, error) {
	file := "./snapshot.json"
	flag.String("configfile", file, "Configuration file path")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetDefault("debug", false)

	viper.SetEnvPrefix("VRSA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(`.`, `_`))
	viper.AutomaticEnv()

	viper.SetConfigFile(viper.GetString("configfile"))
	viperErr := viper.ReadInConfig()

	if viperErr != nil { // Handle errors reading the config file
		log.Fatal(viperErr)
		panic(viperErr)
	}

	if viper.GetBool("debug") {
		fmt.Println(viper.AllSettings())
	}

	c := &Configuration{}

	err := viper.Unmarshal(c)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
		return nil, fmt.Errorf("unable to decode into config struct, %v", err)
	}

	return c, nil
}
