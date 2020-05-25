package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/saidmithilesh/hodor/helpers"
)

// Exported contstants
const (
	TimeUnitSecond  = "S"
	TimeUnitMinute  = "M"
	TimeUnitHour    = "H"
	TimeNil         = time.Duration(0) * time.Second
	CFLevelGateway  = "gateway"
	CFLevelEndpoint = "endpoint"
)

var portRegex = regexp.MustCompile(`:[0-9]+$`)
var durationRegex = regexp.MustCompile(`([0-9]+)(S|s|M|m|H|h){1}`)
var methodsRegex = regexp.MustCompile(`(?i)(^GET$|^PUT$|^POST$|^DELETE$|^OPTIONS$|^PATCH$|^HEAD$)`)

// Conf is a globally accessible singleton instance of type Config.
// It is used by all modules that need to utilise the configuration.
var Conf Config
var once sync.Once

// Config encapsulates the entire application wide configuration.
// Currently it serves the only purpose of wrapping the gateway
// config inside itself. It has been included in the system keeping
// in mind the future need to include more configuration that is
// outside the scope of the gateway's configuration
// TODO: Update the configuration documentation
type Config struct {
	Initialised      bool
	ValidationFailed bool
	ConfigFilePath   string
	Gateway          GatewayConfig `yaml:"gateway"`
}

// GatewayConfig is the parent struct encapsulating all the
// configuration required for the functioning of Hodor.
// When Hodor reads config from the YAML file, it creates an
// instance of this struct and loads the config values into it.
// This instance is then passed around to all modules requiring
// the config to function properly.
type GatewayConfig struct {
	ID          uint   `yaml:"id"`
	InstanceID  uint   `yaml:"instance_id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Port        string `yaml:"port"`

	// TLS configuration
	EnableTLS       bool   `yaml:"enable_TLS"`
	TLSCertFilePath string `yaml:"TLS_cert"`
	TLSKeyFilePath  string `yaml:"TLS_key"`

	// Logging configuration
	LogLevel               string `yaml:"log_level"`
	LogOutput              string `yaml:"log_output"`
	LogCredentialsFilePath string `yaml:"log_credentials"`

	// Gateway wide rate limiting
	RateLimit RateLimiterConfig `yaml:"rate_limit"`

	// GatewayWide CORS
	CORS CORSConfig `yaml:"cors"`

	Endpoints []EndpointConfig `yaml:"endpoints"`
}

// RateLimiterConfig encapsulates the configuration for rate limiting. It
// represents the number of requests allowed within a window of time and the
// penalty to be levied if a user exceeds the specified rate limit.
// Both Window and Penalty are durations of time. To allow users to enter them
// in a more human readable manner, they are represented in a
// '<length><unit_of_time>' format. Ex: '1S' or '10M' or '2H'. Allowed units of
// time are seconds, minutes and hours. The length of the duration can only be
// represented as integers.
// When the config is loaded into memory, these duration strings are immediately
// parsed and broken down into their individual components of length and unit
// and then converted to a unified duration format to allow faster operations.
type RateLimiterConfig struct {
	Enabled       bool   `yaml:"enable"`
	Requests      uint   `yaml:"requests"`
	WindowString  string `yaml:"window"`
	PenaltyString string `yaml:"penalty"`

	// Window and Penalty strings converted into time.Duration
	WindowDuration  time.Duration
	PenaltyDuration time.Duration
	PenaltyEnabled  bool
}

// CORSConfig struct encapsulates the config for enabling CORS
// on an API Wide level as well as per endpoint level
type CORSConfig struct {
	Enabled        bool     `yaml:"enable"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedDomains []string `yaml:"allowed_domains"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	ExposedHeaders []string `yaml:"exposed_headers"`
}

// EndpointConfig struct encapsulates the configuration required
// for each API endpoint individually. Hodor allows for
// endpoints to override the default gateway wide configuration
// for rate limiting and CORS
type EndpointConfig struct {
	ID          uint              `yaml:"id"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	Method      string            `yaml:"method"`
	Path        string            `yaml:"path"`
	RateLimit   RateLimiterConfig `yaml:"rate_limit"`
	CORS        CORSConfig        `yaml:"cors"`
	Middleware  []string          `yaml:"middleware"`
	Backend     string            `yaml:"backend"`
}

// scan method accepts the absolute path to the configuration file
// and reads the content of the file into a byte array and returns it
func (conf *Config) scan(configFilePath string) []byte {
	filecontent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Error while trying to read the config file %#v", err)
	}
	return filecontent
}

func (conf *Config) parse(filecontent []byte) {
	err := yaml.Unmarshal(filecontent, &conf)
	if err != nil {
		log.Fatalf("Error while parsing the config file content into config struct, %#v", err)
	}
	conf.Initialised = true
}

func (conf *Config) validate() {
	conf.Gateway.validate(conf)
	if conf.ValidationFailed {
		log.Print("\n\n")
		log.Fatalln("InvalidConfigError :: One or more validation rules failed. Please fix the above errors in the configuration file before continuing.")
	}
}

func (gc *GatewayConfig) validate(c *Config) {
	gc.validateName(c)
	gc.validatePort(c)
	gc.validateTLS(c)
	gc.RateLimit.validate(CFLevelGateway, c, nil)

	for _, endpoint := range gc.Endpoints {
		endpoint.validate(c)
	}
}

// @param c takes in the configuration
func (gc *GatewayConfig) validateName(c *Config) {
	if gc.Name == "" {
		log.Printf("\t - Error.InvalidGatewayName :: Invalid value '%s' provided. The name field will be used to uniquely identify the gateway in logs and monitoring systems. Please provide a valid name\n", gc.Name)
		c.ValidationFailed = true
	}
}

func (gc *GatewayConfig) validatePort(c *Config) {
	if !portRegex.MatchString(gc.Port) {
		log.Printf("\t - Error.InvalidPort :: Invalid value '%s' provided. Please provide a valid port number in the format ':<portnumber>' to run the gateway\n", gc.Port)
		c.ValidationFailed = true
	}
}

func (gc *GatewayConfig) validateTLS(c *Config) {
	if !gc.EnableTLS {
		return
	}

	gc.TLSCertFilePath = helpers.FilePathHelper.GetFullPath(gc.TLSCertFilePath)
	gc.TLSKeyFilePath = helpers.FilePathHelper.GetFullPath(gc.TLSKeyFilePath)

	if !helpers.FilePathHelper.IsValidPath(gc.TLSCertFilePath) {
		log.Printf("\t - Error.InvalidCertPath :: Invalid filepath '%s' provided for TLS certificate file\n", gc.TLSCertFilePath)
		c.ValidationFailed = true
	}

	if !helpers.FilePathHelper.IsValidPath(gc.TLSKeyFilePath) {
		log.Printf("\t - Error.InvalidKeyPath :: Invalid filepath '%s' provided for TLS key file\n", gc.TLSKeyFilePath)
		c.ValidationFailed = true
	}
}

func (rl *RateLimiterConfig) validate(rlType string, c *Config, e *EndpointConfig) {
	if !rl.Enabled {
		return
	}

	var rlTypeString string
	if rlType == CFLevelGateway {
		rlTypeString = "the gateway"
	} else {
		rlTypeString = fmt.Sprintf("endpoint '%s'", e.Name)
	}

	if rl.Requests == 0 {
		log.Printf("\t - Error.InvalidNumberOfRequests :: Invalid value '%d' provided for rate limiter's allowed number of requests for %s. Please provide a valid integer denoting the number of requests to allow in the specified time window or set the enable flaf to false.\n", rl.Requests, rlTypeString)
		c.ValidationFailed = true
		rl.Enabled = false
	}

	if !durationRegex.MatchString(rl.WindowString) {
		log.Printf("\t - Error.InvalidRateLimitWindow :: Invalid value '%s' provided for rate limiter window for %s. Please provide a valid string of format <window_length><time_unit>. Ex: 10S or 2M or 1H or set the enable flag to false.\n", rl.WindowString, rlTypeString)
		c.ValidationFailed = true
		rl.Enabled = false
	}

	if !durationRegex.MatchString(rl.PenaltyString) && rl.PenaltyString != "-" {
		log.Printf("\t - Error.InvalidRateLimitPenalty :: Invalid value '%s' provided for rate limiter penalty for %s. Please provide a valid string of format <window_length><time_unit>. Ex: 10S or 2M or 1H OR provide a '-' to ignore penalty.\n", rl.PenaltyString, rlTypeString)
		c.ValidationFailed = true
		rl.Enabled = false
	}
}

func (e *EndpointConfig) validate(c *Config) {
	e.validateName(c)
	e.validateMethod(c)
	e.validatePath(c)
	e.validateBackend(c)
	e.RateLimit.validate(CFLevelEndpoint, c, e)
}

func (e *EndpointConfig) validateName(c *Config) {
	if e.Name == "" {
		log.Printf("\t - Error.InvalidEndpointName :: Invalid value '%s' provided. The name field will be used to uniquely identify your endpoints in logs and monitoring systems. Please provide a valid name\n", e.Name)
		c.ValidationFailed = true
	}
}

func (e *EndpointConfig) validateMethod(c *Config) {
	if !methodsRegex.MatchString(e.Method) {
		log.Printf("\t - Error.InvalidMethod :: Invalid value '%s' provided for endpoint %s. Please provide a valid method field. Ex. GET, PUT, POST, DELETE, PATCH, OPTIONS (case insensitive)\n", e.Method, e.Name)
		c.ValidationFailed = true
	}
}

func (e *EndpointConfig) validatePath(c *Config) {
	if e.Path == "" {
		log.Printf("\t - Error.InvalidPath :: Invalid value '%s' provided for endpoint %s. Please provide a valid path.\n", e.Path, e.Name)
		c.ValidationFailed = true
	}
}

func (e *EndpointConfig) validateBackend(c *Config) {
	if e.Path == "" {
		log.Printf("\t - Error.InvalidBackend :: Invalid value '%s' provided for endpoint %s. Please provide a valid http or https backend.\n", e.Backend, e.Name)
		c.ValidationFailed = true
	}
}

// optimise method performs an in-place modification of the config
// instance and optimises it to allow faster operations.
// It converts the configuration from a human readable format to
// one that is easier for the machine to interpret and operate on.
// TODO: Update the documentation
func (conf *Config) optimise() {
	conf.Gateway.optimise(conf)
}

func (gc *GatewayConfig) optimise(c *Config) {
	gc.RateLimit.optimise(CFLevelGateway, c)
	for i, ep := range gc.Endpoints {
		ep.RateLimit.optimise(CFLevelEndpoint, c)
		// to maintain consistency with method names provided by net/http package
		ep.Method = strings.ToUpper(ep.Method)
		gc.Endpoints[i] = ep
	}
}

func (rl *RateLimiterConfig) optimise(rlType string, c *Config) {
	if !rl.Enabled {
		return
	}

	rl.WindowString = strings.ToUpper(rl.WindowString)
	rl.PenaltyString = strings.ToUpper(rl.PenaltyString)

	rl.WindowDuration = stringToDuration(rl.WindowString)
	rl.PenaltyDuration = stringToDuration(rl.PenaltyString)

	// if the penalty duration is 0 seconds,
	// set the penalty enabled flag to false
	if rl.PenaltyDuration != TimeNil {
		rl.PenaltyEnabled = true
	}
}

func stringToDuration(s string) time.Duration {
	// * Users can provide a '-' if they do not wish to levy
	// * any penalty on the clients for exceeding rate limits.
	// * In those cases, ignore and return a "0 second" duration.
	if s == "-" {
		return TimeNil
	}

	matches := durationRegex.FindStringSubmatch(s)
	if len(matches) == 0 {
		log.Fatalln("Error while parsing duration string")
	}

	t, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Fatalln("Error while parsing duration string")
	}

	switch matches[2] {
	case TimeUnitSecond:
		return time.Duration(t) * time.Second

	case TimeUnitMinute:
		return time.Duration(t) * time.Minute

	case TimeUnitHour:
		return time.Duration(t) * time.Hour

	default:
		return TimeNil
	}
}

// ResolveConfigFilePath loads the API Gateway config from a JSON file. It expects
// the filepath to be provided using the "-config" flag when the application is
// started. If no such flag is provided, it assumes a default path
// './config.json'. If a config file is not found at the provided filepath, it
// logs a Fatal error and the program exits.
func ResolveConfigFilePath() string {
	configFilePath := flag.String(
		"config",
		"./config.yml",
		"Absolute or relative path of the API Gateway configuration file",
	)

	flag.Parse()

	// If a relative path is provided, resolve it to an absolute path
	absConfigPath := helpers.FilePathHelper.GetFullPath(*configFilePath)

	// Raise a Fatal error if the filepath is not valid
	if !helpers.FilePathHelper.IsValidPath(absConfigPath) {
		log.Fatalf(
			"Invalid filepath [%s] provided for -config. Please make sure the file exists and is valid.",
			*configFilePath,
		)
	}

	return absConfigPath
}

// LoadConfig function loads the configuration from the config filepath
// provided by the user. It performs 5 crucial steps:
// 1. Resolve the config filepath provided using the -config flag.
// 2. Scan the file and convert it into a byte slice
// 3. Parse the byte slice into an instance of type Config
// 4. Validate the values loaded into the instance
// 5. Optimise the instance
func LoadConfig() Config {
	once.Do(func() {
		configPath := ResolveConfigFilePath()
		filecontent := Conf.scan(configPath)
		Conf.parse(filecontent)
		Conf.validate()
		Conf.optimise()
	})
	return Conf
}

// GetConfig function returns the global instance of type Config
// If it hasn't already been initialised/loaded from file, the
// Load() function is called to do so and return the config
func GetConfig() Config {
	if !Conf.Initialised {
		return LoadConfig()
	}
	return Conf
}
