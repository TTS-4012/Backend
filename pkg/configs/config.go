package configs

import (
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	Conf *OContestConf
)

type OContestConf struct {
	Postgres SectionPostgres `yaml:"postgres"`
	Mongo    SectionMongo    `yaml:"mongo"`
	JWT      SectionJWT      `yaml:"jwt"`
	SMTP     SectionSMTP     `yaml:"smtp"`
	Log      SectionLog      `yaml:"log"`
	Server   SectionServer   `yaml:"server"`
	AESKey   string          `yaml:"AESKey"`
	Auth     SectionAuth     `yaml:"auth"`
}

type SectionLog struct {
	Level        string `yaml:"level"`
	ReportCaller bool   `yaml:"report_caller"`
}

type SectionPostgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type SectionMongo struct {
	Address    string `yaml:"address"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type SectionJWT struct {
	Secret string `yaml:"secret"`
}

type SectionSMTP struct {
	From     string `yaml:"from"`
	Password string `yaml:"password"`
}

type SectionAuth struct {
	Duration SectionAuthDuration `yaml:"duration"`
}

type SectionAuthDuration struct {
	AccessToken  time.Duration `yaml:"accesstoken"`
	RefreshToken time.Duration `yaml:"refreshtoken"`
	VerifyEmail  time.Duration `yaml:"verifyemail"`
}

type SectionServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func getElements(path string, ref reflect.Type) []string {

	var basePath string
	if path != "" {
		basePath = path + "."
	}

	var ans []string
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		fieldPath := strings.ToLower(basePath + field.Tag.Get("yaml"))
		if field.Type.Kind() == reflect.Struct {
			ans = append(ans, getElements(fieldPath, field.Type)...)
		} else {
			ans = append(ans, fieldPath)
		}

	}
	return ans
}

func BindEnvVariables() {
	elms := getElements("", reflect.TypeOf(&OContestConf{}).Elem())
	for _, elm := range elms {
		err := viper.BindEnv(elm)
		if err != nil {
			panic(err)
		}
	}
}

// Loads the config
func getConfig() *OContestConf {
	viper.AutomaticEnv()           // reads from env
	viper.SetEnvPrefix("ocontest") // automatically turns to capitalized
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf := &OContestConf{}
	BindEnvVariables()
	err := viper.Unmarshal(conf)
	if err != nil {
		panic("Error on unmarshal " + err.Error())
	}

	return conf
}

func InitConf() {
	Conf = getConfig()
}
