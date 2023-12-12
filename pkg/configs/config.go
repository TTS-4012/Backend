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
	Nats     SectionNats     `yaml:"nats"`
	JWT      SectionJWT      `yaml:"jwt"`
	SMTP     SectionSMTP     `yaml:"smtp"`
	Log      SectionLog      `yaml:"log"`
	Server   SectionServer   `yaml:"server"`
	AESKey   string          `yaml:"AESKey"`
	Auth     SectionAuth     `yaml:"auth"`
	MinIO    SectionMinIO    `yaml:"minio"`
	Judge    SectionJudge    `yaml:"judge"`
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
	Address  string `yaml:"address"`
	Database string `yaml:"database"`
}

type SectionNats struct {
	Url               string        `yaml:"url"`
	Subject           string        `yaml:"subject"`
	ReplyTimeout      time.Duration `yaml:"reply_timeout"`
	Queue             string        `yaml:"queue"`
	SubscribeChanSize int           `yaml:"subscribe_chan_size"`
}

type SectionJWT struct {
	Secret string `yaml:"secret"`
}

type SectionSMTP struct {
	From     string `yaml:"from"`
	Password string `yaml:"password"`
	Enabled  bool   `yaml:"enabled"`
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

type SectionMinIO struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"accesskey"`
	SecretKey string `yaml:"secretkey"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	Secure    bool   `yaml:"secure"`
}

type SectionJudge struct {
	Nats SectionNats `yaml:"nats"`
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
