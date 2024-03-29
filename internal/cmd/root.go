package cmd

import (
	"errors"
	"fmt"
	"github.com/gopenapi/gopenapi/internal/pkg/openapi"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
)

const version = "0.0.3"

var rootCmd = &cobra.Command{
	Use:     "gopenapi",
	Short:   "gopenapi",
	Long:    `Gopenapi use javascript to extend and simplify openapi sepc`,
	Version: version,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		confPath := cmd.Flag("config").Value.String()
		_, err := os.Lstat(confPath)
		if err != nil {
			if os.IsNotExist(err) {
				err = ioutil.WriteFile(confPath, []byte(defaultConfig), os.ModePerm)
				if err != nil {
					err = fmt.Errorf("wirte default config file err: %w", err)
					return err
				}
			} else {
				return err
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		confFile := cmd.Flag("config").Value.String()
		input := cmd.Flag("input").Value.String()
		output := cmd.Flag("output").Value.String()

		if input == "" || output == "" {
			return errors.New("invalid input or output, please type 'gopenapi -h' to get help")
		}
		o, err := openapi.NewOpenApi("go.mod", confFile)
		if err != nil {
			return err
		}
		inputBs, err := ioutil.ReadFile(input)
		if err != nil {
			return err
		}

		format := openapi.Yaml
		if path.Ext(output) == ".json" {
			format = openapi.Json
		}
		outputYaml, err := o.CompleteYaml(string(inputBs), format)
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
