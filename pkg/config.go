package pkg

import (
	"log"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

var (
	Conf *OContestConf
)

type OContestConf struct {
	Port int        `yaml:"Port"`
	Log  SectionLog `yaml:"log"`
}

type SectionLog struct {
	Level        string `yaml:"level"`
	ReportCaller bool   `yaml:"report_caller"`
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
	viper.SetConfigType("yaml")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/ocontest/")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Println("Error on reading config", err)
	}

	viper.AutomaticEnv()           // reads from env
	viper.SetEnvPrefix("ocontest") // automatically turns to capitalized
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf := &OContestConf{}
	BindEnvVariables()
	err = viper.Unmarshal(conf)
	if err != nil {
		panic("Error on unmarshal " + err.Error())
	}

	return conf
}

func InitConf() {
	Conf = getConfig()
	initLog(Conf)
	Log.Info("Config loaded successfully")
}