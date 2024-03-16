package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	AuthToken string `yaml:"authToken"`
	IpAddress string `yaml:"ipAddress"`
	Port      string `yaml:"port"`
	Commands  []struct {
		Name string `yaml:"name"`
		Uri  string `yaml:"uri"`
		Cmd  string `yaml:"cmd"`
	}
}

type Command struct {
	cmd   string
	name  string
	token string
}

type Response struct {
	CommandName string `json:"commandName"`
	Output      string `json:"output"`
}

func (command *Command) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var response Response
	log := fmt.Sprintf("Request from %s, cmdname: %s", r.RemoteAddr, command.name)
	zap.S().Info(log)
	if r.Header.Get("Authorization-token") != command.token {
		zap.S().Error("Authorization error")
		msg := []byte(`{"message":"Unauthorized"}`)
		http.Error(w, string(msg), http.StatusUnauthorized)
		return
	}

	response.CommandName = command.name

	output, err := exec.Command("sh", "-c", command.cmd).CombinedOutput()
	if err != nil {
		log = fmt.Sprintf("Command \"%s\", execution error: %s", command.name, err)
		zap.S().Error(log)
		msg := fmt.Sprintf("{\"message\": \"Execution error. Command: %s\"}", command.name)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	log = fmt.Sprintf("Successfully executed command \"%s\", output: %s", command.name, output)
	zap.S().Debug(log)
	// log = fmt.Sprintf("Successfully executed command \"%s\"", command.name)
	// zap.S().Info(log)

	msg := fmt.Sprintf("{\"message\": \"Successfully executed. Command: %s\"}", command.name)
	fmt.Fprint(w, msg)
}

func loadConfig(filename string, settings *Config) (err error) {
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		return
	}
	file, err := os.ReadFile(filename)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(file, settings)
	return
}

func getLogger() *zap.Logger {
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout"],
	  "errorOutputPaths": ["stderr"],
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)
	var cfg zap.Config
	_ = json.Unmarshal(rawJSON, &cfg)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	cfg.EncoderConfig.TimeKey = "ts"
	logger := zap.Must(cfg.Build())
	return logger
}

func main() {
	var (
		config Config
	)
	log := getLogger()
	defer log.Sync()
	undo := zap.ReplaceGlobals(log)
	defer undo()

	log.Debug("Reading config")
	err := loadConfig("config.yaml", &config)
	if err != nil {
		log.Fatal(err.Error())
	}
	for _, command := range config.Commands {
		handler := &Command{cmd: command.Cmd, name: command.Name, token: config.AuthToken}
		http.Handle(command.Uri, handler)
		log.Debug(fmt.Sprintf("Added command %s", command.Name))
	}

	log.Info(fmt.Sprintf("Starting server on %s:%s", config.IpAddress, config.Port))
	if err := http.ListenAndServe(fmt.Sprintf("%s:%s", config.IpAddress, config.Port), nil); err != nil {
		log.Fatal(err.Error())
	}
}
