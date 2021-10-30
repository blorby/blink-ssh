package implementation

import (
	"errors"
	"fmt"
	"github.com/blinkops/blink-sdk/plugin"
	"github.com/blinkops/blink-sdk/plugin/actions"
	"github.com/blinkops/blink-sdk/plugin/config"
	"github.com/blinkops/blink-sdk/plugin/connections"
	description2 "github.com/blinkops/blink-sdk/plugin/description"
	log "github.com/sirupsen/logrus"
	"path"
	"strconv"
	"time"
)

const (
	hostKey    = "host"
	timeoutKey = "timeout"
	portKey    = "port"
	commandKey = "command"
)

type ActionHandler func(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error)

type SSHPlugin struct {
	description      plugin.Description
	actions          []plugin.Action
	supportedActions map[string]ActionHandler
}

func (p *SSHPlugin) Describe() plugin.Description {
	log.Debug("Handling Describe request!")
	return p.description
}

func (p *SSHPlugin) GetActions() []plugin.Action {
	log.Debug("Handling GetActions request!")
	return p.actions
}

func (p *SSHPlugin) ExecuteAction(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) (*plugin.ExecuteActionResponse, error) {
	log.Debugf("Executing action: %v\n Context: %v", *request, ctx.GetAllContextEntries())

	actionHandler, ok := p.supportedActions[request.Name]
	if !ok {
		return nil, errors.New("action is not supported: " + request.Name)
	}

	resultBytes, err := actionHandler(ctx, request)
	if len(resultBytes) > 0 && resultBytes[len(resultBytes)-1] == '\n' {
		resultBytes = resultBytes[:len(resultBytes)-1]
	}

	if err != nil {
		msg := fmt.Sprintf(" error: %v, output: %v", err, string(resultBytes))
		log.Errorf("%s", msg)
		return &plugin.ExecuteActionResponse{
			ErrorCode: 1,
			Result:    []byte(msg),
		}, nil
	}
	return &plugin.ExecuteActionResponse{
		ErrorCode: 0,
		Result:    resultBytes,
	}, nil
}

func (p *SSHPlugin) TestCredentials(_ map[string]connections.ConnectionInstance) (*plugin.CredentialsValidationResponse, error) {
	return &plugin.CredentialsValidationResponse{
		AreCredentialsValid:   true,
		RawValidationResponse: []byte("credentials validation is not supported on this plugin :("),
	}, nil
}

func NewSSHPlugin(rootPluginDirectory string) (*SSHPlugin, error) {

	pluginConfig := config.GetConfig()

	description, err := description2.LoadPluginDescriptionFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	loadedConnections, err := connections.LoadConnectionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.PluginDescriptionFilePath))
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d connections from disk", len(loadedConnections))
	description.Connections = loadedConnections

	actionsFromDisk, err := actions.LoadActionsFromDisk(path.Join(rootPluginDirectory, pluginConfig.Plugin.ActionsFolderPath))
	if err != nil {
		return nil, err
	}

	supportedActions := map[string]ActionHandler{
		"execute": executeSSH,
	}

	return &SSHPlugin{
		description:      *description,
		actions:          actionsFromDisk,
		supportedActions: supportedActions,
	}, nil
}

func executeSSH(ctx *plugin.ActionContext, request *plugin.ExecuteActionRequest) ([]byte, error) {

	host, ok := request.Parameters[hostKey]
	if !ok || host == "" {
		return nil, errors.New("no host parameter provided for execution")
	}

	timeoutStr, ok := request.Parameters[timeoutKey]
	if !ok {
		timeoutStr = "60"
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, errors.New("timeout should be a number, got " + timeoutStr)
	}
	timeOutDuration := time.Duration(timeout) * time.Second

	portStr, ok := request.Parameters[portKey]
	if !ok {
		portStr = "22"
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, errors.New("port should be a number, got " + portStr)
	}

	command, ok := request.Parameters[commandKey]
	if !ok || command == "" {
		return nil, errors.New("no command parameter provided for execution")
	}

	credentials, err := ctx.GetCredentials("ssh")
	if err != nil {
		return nil, errors.New("missing ssh connection")
	}

	key, ok := credentials["key"].(string)
	if !ok || key == "" {
		return nil, errors.New("missing ssh key")
	}

	user, ok := credentials["username"].(string)
	if !ok || user == "" {
		return nil, errors.New("missing ssh username")
	}

	passphrase, ok := credentials["passphrase"].(string)
	if !ok {
		passphrase = ""
	}

	plugin := Plugin{
		Config: Config{
			Key:        key,
			Username:   user,
			Passphrase: passphrase,
			Host:       host,
			Port:       port,
			Timeout:    timeOutDuration,
			Script:     createCommand(command),
		},
	}

	log.Infof("About to do ssh call for the next command: %v", plugin.Config.Script)
	output, err := plugin.Exec()
	log.Infof("Got response back, output: %v, err: %v", output, err)

	return []byte(output), err

}

func createCommand(cmd string) []string {
	commands := make([]string, 0)
	commands = append(commands, cmd)
	commands = append(commands, "BLINK_SSH_PREV_COMMAND_EXIT_CODE=$? ; if [ $BLINK_SSH_PREV_COMMAND_EXIT_CODE -ne 0 ]; then exit $BLINK_SSH_PREV_COMMAND_EXIT_CODE; fi;")
	return commands
}
