package main

import (
	cc "github.com/faelmori/cleandgo/cmd/cli"
	gl "github.com/faelmori/cleandgo/logger"
	vs "github.com/faelmori/cleandgo/version"
	"github.com/spf13/cobra"

	"os"
	"strings"
)

type CleandGO struct {
	parentCmdName string
	printBanner   bool
}

func (m *CleandGO) Alias() string { return "" }
func (m *CleandGO) ShortDescription() string {
	return "CleandGO is a minimalistic backend service with Go."
}
func (m *CleandGO) LongDescription() string {
	return `CleandGO: A minimalistic backend service with Go.`
}
func (m *CleandGO) Usage() string {
	return "cleandgo [command] [args]"
}
func (m *CleandGO) Examples() []string {
	return []string{"cleandgo start -p ':8080' -b '0.0.0.0' -n 'MyService' -d"}
}
func (m *CleandGO) Active() bool {
	return true
}
func (m *CleandGO) Module() string {
	return "cleandgo"
}
func (m *CleandGO) Execute() error { return m.Command().Execute() }
func (m *CleandGO) Command() *cobra.Command {
	gl.Log("debug", "Starting CleandGO CLI...")

	var rtCmd = &cobra.Command{
		Use:     m.Module(),
		Aliases: []string{m.Alias()},
		Example: m.concatenateExamples(),
		Version: vs.GetVersion(),
		Annotations: cc.GetDescriptions([]string{
			m.LongDescription(),
			m.ShortDescription(),
		}, m.printBanner),
	}

	rtCmd.AddCommand(cc.ServiceCmdList()...)
	rtCmd.AddCommand(vs.CliCommand())

	// Set usage definitions for the command and its subcommands
	setUsageDefinition(rtCmd)
	for _, c := range rtCmd.Commands() {
		setUsageDefinition(c)
		if !strings.Contains(strings.Join(os.Args, " "), c.Use) {
			if c.Short == "" {
				c.Short = c.Annotations["description"]
			}
		}
	}

	return rtCmd
}
func (m *CleandGO) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *CleandGO) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}

func RegX() *CleandGO {
	var printBannerV = os.Getenv("GOBEMIN_PRINT_BANNER")
	if printBannerV == "" {
		printBannerV = "true"
	}

	return &CleandGO{
		printBanner: strings.ToLower(printBannerV) == "true",
	}
}
