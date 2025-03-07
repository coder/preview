package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/coder/serpent"
)

func (r *RootCmd) TerraformPlan() *serpent.Command {
	cmd := &serpent.Command{
		Use:   "plan",
		Short: "Runs `terraform init -upgrade` and `terraform plan`, saving the output.",
		// This command is mainly for developing the preview tool.
		Hidden: true,
		Handler: func(i *serpent.Invocation) error {
			ctx := i.Context()

			cmd := exec.CommandContext(ctx, "terraform", "init", "-upgrade")
			cmd.Stdin = i.Stdin
			cmd.Stdout = i.Stdout
			cmd.Stderr = i.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("terraform init: %w", err)
			}

			cmd = exec.CommandContext(ctx, "terraform", "plan", "-out", "out.plan")
			cmd.Stdin = i.Stdin
			cmd.Stdout = i.Stdout
			cmd.Stderr = i.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("terraform plan: %w", err)
			}

			var buf bytes.Buffer
			cmd = exec.CommandContext(ctx, "terraform", "show", "-json", "out.plan")
			cmd.Stdin = i.Stdin
			cmd.Stdout = &buf
			cmd.Stderr = i.Stderr

			if err := cmd.Run(); err != nil {
				_, _ = cmd.Stdout.Write(buf.Bytes())
				return fmt.Errorf("terraform show: %w", err)
			}

			if !cmd.ProcessState.Success() {
				return fmt.Errorf("terraform show not successful: %v", cmd.ProcessState)
			}

			var indented bytes.Buffer
			_ = json.Indent(&indented, buf.Bytes(), "", "  ")
			_ = os.WriteFile("plan.json", indented.Bytes(), 0644)
			return nil
		},
	}

	return cmd
}
