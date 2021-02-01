package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"github.com/zbysir/gopenapi/internal/pkg/openapi"
	"io/ioutil"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "gopenapi",
	Short: "gopenapi",
	Long:  `Gopenapi is progressive generator to generate openapi spec from golang source`,
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
}

func Execute() error {
	rootCmd.Flags().String("config", "gopenapi.conf.js", "Specify the configuration file to be used")
	rootCmd.Flags().String("input", "", "Specify the source file in yaml format")
	rootCmd.Flags().String("output", "", "Specify the output file path")

	return rootCmd.Execute()
}
