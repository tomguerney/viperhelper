package viperhelper

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/tomguerney/printer"
)

var h = &Viperhelper{}

func init() {
	setPrinterStencils()
}

// Rule is a validation rule
type Rule func(i interface{}) (string, bool)

type option struct {
	required    bool
	name        string
	flag        string
	description string
	value       interface{}
	envPrefix   string
	errMsg      string
	rule        Rule
}

func (o *option) toMap() map[string]string {
	configfile := fmt.Sprintf("%v: <%v>", o.name, o.name)
	envvar := strings.ToUpper(o.name)
	if o.envPrefix != "" {
		envvar = fmt.Sprintf("%v_%v", strings.ToUpper(h.envPrefix), envvar)
	}
	return map[string]string{
		"required":    strconv.FormatBool(o.required),
		"name":        o.name,
		"flag":        o.flag,
		"description": o.description,
		"value":       fmt.Sprint(o.value),
		"errMsg":      o.errMsg,
		"configfile":  configfile,
		"envvar":      envvar,
	}
}

// Viperhelper is a utility wrapper for Viper
type Viperhelper struct {
	options   []*option
	envPrefix string
	viper     *viper.Viper
}

// GetViper
func GetViper() *viper.Viper {
	return h.GetViper()
}

func (h *Viperhelper) GetViper() *viper.Viper {
	if h.viper == nil {
		panic("viper not set")
	}
	return h.viper
}

func SetViper(viper *viper.Viper) {
	h.SetViper(viper)
}

func (h *Viperhelper) SetViper(viper *viper.Viper) {
	h.viper = viper
}

func SetEnvPrefix(prefix string) {
	h.SetEnvPrefix((prefix))
}

func (h *Viperhelper) SetEnvPrefix(prefix string) {
	h.envPrefix = prefix
	h.GetViper().SetEnvPrefix((prefix))
	for _, option := range h.options {
		option.envPrefix = prefix
	}
}

func AddEnv(name, description, flag string, required bool, rule Rule) {
	h.AddEnv(name, description, flag, required, rule)
}

func (h *Viperhelper) AddEnv(name, description, flag string, required bool, rule Rule) {
	h.options = append(h.options, &option{
		required:    required,
		name:        name,
		description: description,
		flag:        flag,
		rule:        rule,
		envPrefix:   h.envPrefix,
	})
}

func Validate() error {
	return h.Validate()
}

func (h *Viperhelper) Validate() error {
	omitted := h.getOmittedEnvs()
	invalid := h.getInvalidEnvs()
	if len(omitted) > 0 {
		h.printOmittedOptions(omitted)
	}
	if len(invalid) > 0 {
		h.printInvalidEnvs(invalid)
	}
	if len(omitted) > 0 || len(invalid) > 0 {
		log.Info().Msg("configuration invalid")
		return errors.New("configuration invalid")
	}
	return nil
}

func (h *Viperhelper) getOmittedEnvs() (omitted []*option) {
	for _, option := range h.options {
		if option.required && !h.GetViper().IsSet(option.name) {
			omitted = append(omitted, option)
		}
	}
	return
}

func (h *Viperhelper) getInvalidEnvs() (invalid []*option) {
	for _, o := range h.options {
		if i := h.GetViper().Get(o.name); i != nil && o.rule != nil {
			if errMsg, ok := o.rule(i); !ok {
				o.value = i
				o.errMsg = errMsg
				invalid = append(invalid, o)
			}
		}
	}
	return
}

func (h *Viperhelper) printOmittedOptions(omitted []*option) {
	printer.Feed()
	printer.Out(printer.Color("Some required configuration options haven't been set: ", printer.Red))
	optionMap := []map[string]string{}
	for _, option := range omitted {
		optionMap = append(optionMap, option.toMap())
	}
	printer.Feed()
	printer.TableStencil("omitted-options", optionMap)
	printer.Feed()
	printer.Out("You can either set them as flags, environment variables, or in the config file:")
	printer.Feed()
	printer.TableStencil("option-choices", optionMap)
	printer.Feed()
}

func (h *Viperhelper) printInvalidEnvs(invalid []*option) {
	printer.Feed()
	printer.Out(printer.Color("Some configuration options are invalid: ", printer.Red))
	optionMap := []map[string]string{}
	for _, option := range invalid {
		optionMap = append(optionMap, option.toMap())
	}
	printer.Feed()
	printer.TableStencil("invalid-options", optionMap)
	printer.Feed()
}

func setPrinterStencils() {
	printer.AddTableStencil(
		"omitted-options",
		[]string{"Option", "Description"},
		[]string{"name", "description"},
		nil,
	)

	printer.AddTableStencil(
		"invalid-options",
		[]string{"Option", "Value", "Error"},
		[]string{"name", "value", "errMsg"},
		nil,
	)

	printer.AddTableStencil(
		"option-choices",
		[]string{"Option", "Flag", "Environment", "Config File"},
		[]string{"name", "flag", "envvar", "configfile"},
		map[string]string{
			"flag":       printer.Green,
			"envvar":     printer.Green,
			"configfile": printer.Green,
		},
	)
}
