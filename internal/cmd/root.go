package cmd

import (
	"errors"
	"github.com/gopenapi/gopenapi/internal/pkg/openapi"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

const version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:     "gopenapi",
	Short:   "gopenapi",
	Long:    `Gopenapi is progressive generator to generate openapi spec from golang source`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		confFile := cmd.Flag("config").Value.String()
		input := cmd.Flag("input").Value.String()
		output := cmd.Flag("output").Value.String()

		if input == "" || output == "" {
			return errors.New("invalid input or output")
		}
		o, err := openapi.NewOpenApi("go.mod", confFile)
		if err != nil {
			return err
		}
		inputBs, err := ioutil.ReadFile(input)
		if err != nil {
			return err
		}
		outputYaml, err := o.CompleteYaml(string(inputBs))
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(output, []byte(outputYaml), os.ModePerm)
		if err != nil {
			return err
		}

		return nil
	},
	SilenceUsage: true,
}

func Execute() error {
	rootCmd.Flags().StringP("config", "c", "gopenapi.conf.js", "Specify the configuration file to be used")
	rootCmd.Flags().StringP("input", "i", "", "Specify the source file in yaml format")
	rootCmd.Flags().StringP("output", "o", "", "Specify the output file path")

	return rootCmd.Execute()
}
