package instance

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
	"github.com/valyentdev/ravel/core/daemon"
	"github.com/valyentdev/ravel/core/instance"
	"sigs.k8s.io/yaml"
)

type createOptions struct {
	config string
}

func newCreateInstanceCmd() *cobra.Command {
	var opts createOptions

	createCmd := &cobra.Command{
		Use:   "create <id>",
		Short: "Create a new instance",
		Long: `Create & start a new instance from a given image
The instance spec is defined in a json or yaml file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createInstance(cmd, args[0], opts)
		},
		Args: cobra.ExactArgs(1),
	}

	createCmd.Flags().StringVarP(&opts.config, "config", "c", "", "Config file which contains instance spec")
	createCmd.MarkFlagRequired("config")

	return createCmd
}

func createInstance(cmd *cobra.Command, id string, opt createOptions) error {
	var file []byte
	var err error

	if strings.HasPrefix(opt.config, "http") {
		resp, err := http.Get(opt.config)
		if err != nil {
			return fmt.Errorf("unable to fetch config from URL %s: %w", opt.config, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch config from URL %s: %s", opt.config, resp.Status)
		}

		file, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unable to read config from URL %s: %w", opt.config, err)
		}
	} else {
		file, err = os.ReadFile(opt.config)
		if err != nil {
			return fmt.Errorf("unable to read config file %s: %w", opt.config, err)
		}
	}

	var config instance.InstanceConfig
	if err = yaml.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("unable to unmarshal config file %s: %w", opt.config, err)
	}

	res, err := util.GetDaemonClient(cmd).CreateInstance(cmd.Context(), daemon.InstanceOptions{
		Id:     id,
		Config: config,
	})

	if err != nil {
		return fmt.Errorf("unable to create instance: %w", err)
	}

	fmt.Println(res.Id)
	return nil
}
