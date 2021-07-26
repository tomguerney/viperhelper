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

type AddOptionOptions struct {
	Name        string
	Description string
	Example     string
	Env         string
	Flag        string
	Required    bool
	Rule        Rule
}

type option struct {
	required    bool
	name        string
	flag        string
	description string
	example     string
	env         string
	value       interface{}
	envPrefix   string
	errMsg      string
	rule        Rule
}

func (o *option) toMap() map[string]string {
	configfile := fmt.Sprintf("%v: <%v>", o.name, o.name)
	envvar := strings.ToUpper(o.name)
	if o.env != "" {
		envvar = o.env
	}
	if o.envPrefix != "" {
		envvar = fmt.Sprintf("%v_%v", strings.ToUpper(h.envPrefix), envvar)
	}
	return map[string]string{
		"required":    strconv.FormatBool(o.required),
		"name":        o.name,
		"flag":        o.flag,
		"description": o.description,
		"example":     o.example,
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

func AddOption(opts *AddOptionOptions) {
	h.AddOption(opts)
}

func (h *Viperhelper) AddOption(opts *AddOptionOptions) {
	h.options = append(h.options, &option{
		required:    opts.Required,
		name:        opts.Name,
		description: opts.Description,
		example:     opts.Example,
		flag:        opts.Flag,
		env:         opts.Env,
		rule:        opts.Rule,
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
	printer.UseTableStencil("omitted-options", optionMap)
	printer.Feed()
	printer.Out("You can either set them as flags, environment variables, or in the config file:")
	printer.Feed()
	printer.UseTableStencil("option-choices", optionMap)
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
	printer.UseTableStencil("invalid-options", optionMap)
	printer.Feed()
}

func setPrinterStencils() {
	printer.AddTableStencil(&printer.TableStencil{
		ID:          "omitted-options",
		Headers:     []string{"Option", "Description", "Example"},
		ColumnOrder: []string{"name", "description", "example"},
	})

	printer.AddTableStencil(&printer.TableStencil{
		ID:          "invalid-options",
		Headers:     []string{"Option", "Value", "Error"},
		ColumnOrder: []string{"name", "value", "errMsg"},
	})

	printer.AddTableStencil(&printer.TableStencil{
		ID:          "option-choices",
		Headers:     []string{"Option", "Flag", "Environment", "Config File"},
		ColumnOrder: []string{"name", "flag", "envvar", "configfile"},
		Colors: map[string]string{
			"flag":       printer.Green,
			"envvar":     printer.Green,
			"configfile": printer.Green,
		},
	})

}
