package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func mustBindPFlag(key string, f *flag.Flag) {
	if err := viper.BindPFlag(key, f); err != nil {
		panic(fmt.Sprintf("viper.BindPFlag(%s) failed: %v", key, err))
	}
}

// Init injects the server flags into a cobra command
func Init(cmd *cobra.Command) {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AddConfigPath("/etc/config")
	viper.SetConfigName("server")
	viper.SetConfigType("yaml")

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.SetEnvPrefix("http")
	viper.AutomaticEnv()

	// Server flags
	flagSet := flag.NewFlagSet("server", flag.ExitOnError)
	flagSet.BoolP("debug", "d", false, "Set log level to debug")
	flagSet.IntP("port", "p", 8080, "Port for the http listener")
	flagSet.IntP("securePort", "s", 8443, "Port for the HTTPS listener")
	flagSet.StringP("user", "U", "", "Set BasicAuth user for the http listener")
	flagSet.StringP("pass", "P", "", "Set BasicAuth password for the http listener")
	flagSet.StringP("certFile", "C", "", "File containing the x509 Certificate for HTTPS.")
	flagSet.StringP("keyFile", "K", "", "File containing the x509 private key for HTTPS.")
	flagSet.IntP("request-limit", "l", 0, "Max requests per second per client allowed")
	flagSet.BoolP("pretty-json", "", false, "Pretty print JSON responses")
	flagSet.BoolP("localhost-only", "", false, "wether to bind to 127.0.0.1 or 0.0.0.0")
	flagSet.BoolP("strip-slashes", "S", false, "Strip trailing slashes befofore matching routes")

	cmd.Flags().AddFlagSet(flagSet)
	if err := flagSet.Parse(flag.Args()); err != nil {
		panic(err)
	}

	// viper.BindPFlags(flag.CommandLine)
	mustBindPFlag("debug", flagSet.Lookup("debug"))
	mustBindPFlag("stripSlashes", flagSet.Lookup("strip-slashes"))
	mustBindPFlag("port", flagSet.Lookup("port"))
	mustBindPFlag("securePort", flagSet.Lookup("securePort"))
	mustBindPFlag("auth.user", flagSet.Lookup("user"))
	mustBindPFlag("auth.pass", flagSet.Lookup("pass"))
	mustBindPFlag("certFile", flagSet.Lookup("certFile"))
	mustBindPFlag("keyFile", flagSet.Lookup("keyFile"))
	mustBindPFlag("requestLimit", flagSet.Lookup("request-limit"))
	mustBindPFlag("prettyJSON", flagSet.Lookup("pretty-json"))
	mustBindPFlag("localHostOnly", flagSet.Lookup("localhost-only"))
}
