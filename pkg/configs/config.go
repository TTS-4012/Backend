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
	SQLDB   SectionSQLDB   `yaml:"sql_db"`
	KVStore SectionKVStore `yaml:"kvstore"`
	Mongo   SectionMongo   `yaml:"mongo"`
	JWT     SectionJWT     `yaml:"jwt"`
	SMTP    SectionSMTP    `yaml:"smtp"`
	Log     SectionLog     `yaml:"log"`
	Server  SectionServer  `yaml:"server"`
	AESKey  string         `yaml:"AESKey"`
	Auth    SectionAuth    `yaml:"auth"`
	MinIO   SectionMinIO   `yaml:"minio"`
	Judge   SectionJudge   `yaml:"judge"`
}

type SectionLog struct {
	Level        string `yaml:"level"`
	ReportCaller bool   `yaml:"report_caller"`
}

type SectionSQLDB struct {
	DBType   string          `yaml:"type"`
	ConnUrl  string          `yaml:"conn_url"`
	Postgres SectionPostgres `yaml:"postgres"`
}

type SectionPostgres struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type SectionKVStore struct {
	Type  string       `yaml:"type"`
	Redis SectionRedis `yaml:"redis"`
}

type SectionRedis struct {
	Address string        `yaml:"address"`
	DB      int           `yaml:"db"`
	Timeout time.Duration `yaml:"timeout"`
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
	Host                   string        `yaml:"host"`
	Port                   string        `yaml:"port"`
	GracefulShutdownPeriod time.Duration `yaml:"graceful_shutdown_period"`
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
	c.Server.GracefulShutdownPeriod = viper.GetDuration("server.graceful_shutdown_period")
	c.SQLDB.DBType = viper.GetString("sql_db.type")
	c.SQLDB.ConnUrl = viper.GetString("sql_db.conn_url")
	c.SQLDB.Postgres.Host = viper.GetString("sql_db.postgres.host")
	c.SQLDB.Postgres.Port = viper.GetInt("sql_db.postgres.port")
	c.SQLDB.Postgres.Username = viper.GetString("sql_db.postgres.username")
	c.SQLDB.Postgres.Password = viper.GetString("sql_db.postgres.password")
	c.SQLDB.Postgres.Database = viper.GetString("sql_db.postgres.database")
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
