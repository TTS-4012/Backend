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
	Redis    SectionRedis    `yaml:"redis"`
	Mongo    SectionMongo    `yaml:"mongo"`
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

type SectionRedis struct {
	Address         string        `yaml:"address"`
	DB              int           `yaml:"db"`
	Timeout         time.Duration `yaml:"timeout"`
}

type SectionMongo struct {
	Address  string `yaml:"address"`
	Database string `yaml:"database"`
}

type SectionNats struct {
	Url          string        `yaml:"url"`
	Subject      string        `yaml:"subject"`
	ReplyTimeout time.Duration `yaml:"reply_timeout"`
	Queue        string        `yaml:"queue"`
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
	AccessToken  time.Duration `yaml:"access_token"`
	RefreshToken time.Duration `yaml:"refresh_token"`
	VerifyEmail  time.Duration `yaml:"verify_email"`
}

type SectionServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type SectionMinIO struct {
	Enabled   bool   `yaml:"enabled"`
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	Secure    bool   `yaml:"secure"`
}

type SectionJudge struct {
	EnableRunner bool        `yaml:"enable_runner"` // if it is set, then there is no need for separate runner app. not recommended
	Nats         SectionNats `yaml:"nats"`
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

func AddVariablesWithUnderscore(c *OContestConf) {
	c.Judge.EnableRunner = viper.GetBool("judge.enable_runner")
	c.Judge.Nats.ReplyTimeout = viper.GetDuration("judge.nats.reply_timeout")
	c.MinIO.AccessKey = viper.GetString("minio.access_key")
	c.MinIO.SecretKey = viper.GetString("minio.secret_key")
	c.Auth.Duration.AccessToken = viper.GetDuration("auth.duration.access_token")
	c.Auth.Duration.RefreshToken = viper.GetDuration("auth.duration.refresh_token")
	c.Auth.Duration.VerifyEmail = viper.GetDuration("auth.duration.verify_email")
}

// Loads the config
func getConfig() *OContestConf {
	viper.AutomaticEnv()           // reads from env
	viper.SetEnvPrefix("ocontest") // automatically turns to capitalized
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	conf := &OContestConf{}
	BindEnvVariables()
	AddVariablesWithUnderscore(conf)
	err := viper.Unmarshal(conf)
	if err != nil {
		panic("Error on unmarshal " + err.Error())
	}

	return conf
}

func InitConf() {
	Conf = getConfig()
}
