package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/coder/serpent"
)

func (r *RootCmd) SetEnv() *serpent.Command {
	var (
		vars   []string
		groups []string
	)

	cmd := &serpent.Command{
		Use:   "env",
		Short: "Sets environment variables for terraform plan/apply.",
		Options: []serpent.Option{
			{
				Name:          "vars",
				Description:   "Variables.",
				Flag:          "vars",
				FlagShorthand: "v",
				Default:       "",
				Value:         serpent.StringArrayOf(&vars),
			},
			{
				Name:          "groups",
				Description:   "Groups.",
				Flag:          "groups",
				FlagShorthand: "g",
				Default:       "",
				Value:         serpent.StringArrayOf(&groups),
			},
		},
		Hidden: false,
		Handler: func(i *serpent.Invocation) error {
			for _, val := range vars {
				parts := strings.Split(val, "=")
				if len(parts) != 2 {
					continue
				}
				sum := sha256.Sum256([]byte(parts[0]))
				err := os.Setenv("CODER_PARAMETER_"+hex.EncodeToString(sum[:]), parts[1])
				if err != nil {
					return err
				}
				fmt.Println("CODER_PARAMETER_" + hex.EncodeToString(sum[:]) + "=" + parts[1])
			}

			return nil
		},
	}

	return cmd
}
