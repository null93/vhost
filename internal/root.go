package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/jetrails/proposal-nginx/pkg/vhost"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ParseKeyValueArgs(args []string) vhost.TemplateInput {
	mapping := vhost.TemplateInput{}
	for _, arg := range args {
		parts := strings.Split(arg, "=")
		mapping[parts[0]] = parts[1]
	}
	return mapping
}

func MergeInput(input1, input2 vhost.TemplateInput) vhost.TemplateInput {
	merged := vhost.TemplateInput{}
	for key, value := range input1 {
		merged[key] = value
	}
	for key, value := range input2 {
		merged[key] = value
	}
	return merged
}

func ValidateKeyValueArgs(after int) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, arg := range args[after:] {
			parts := strings.Split(arg, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid key value pair passed as argument %q", arg)
			}
		}
		return nil
	}
}

func ExitWithError(code int, message string) {
	fmt.Printf("\nError: %s\n\n", message)
	os.Exit(code)
}

var RootCmd = &cobra.Command{
	Use:     "vhost",
	Version: "0.0.1",
	Short:   "manage nginx virtual hosts",
}

func initConfig() {
	viper.AddConfigPath("/etc/vhost")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetConfigPermissions(os.FileMode(0644))
	viper.SetDefault("nginx_path", vhost.PATH_NGINX_DIR)
	viper.SetDefault("nginx_available_path", vhost.PATH_NGINX_AVAILABLE_DIR)
	viper.SetDefault("nginx_enabled_path", vhost.PATH_NGINX_ENABLED_DIR)
	viper.SetDefault("templates_path", vhost.PATH_TEMPLATES_DIR)
	viper.SetDefault("checkpoints_path", vhost.PATH_CHECKPOINTS_DIR)
	viper.SafeWriteConfig()
	viper.ReadInConfig()
	vhost.PATH_NGINX_DIR = viper.GetString("nginx_path")
	vhost.PATH_NGINX_AVAILABLE_DIR = viper.GetString("nginx_available_path")
	vhost.PATH_NGINX_ENABLED_DIR = viper.GetString("nginx_enabled_path")
	vhost.PATH_TEMPLATES_DIR = viper.GetString("templates_path")
	vhost.PATH_CHECKPOINTS_DIR = viper.GetString("checkpoints_path")
}

func init() {
	RootCmd.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})
	RootCmd.CompletionOptions.DisableDefaultCmd = true
	cobra.OnInitialize(initConfig)
}
