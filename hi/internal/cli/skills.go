package cli

import (
	"context"
	"fmt"
	"hi/internal/skills"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func CmdSkills() *cobra.Command {
	cmd := &cobra.Command{Use: "skills", Short: "Manage skills"}
	cmd.AddCommand(
		&cobra.Command{Use: "list", Short: "List installed skills", RunE: runSkillsList},
		&cobra.Command{Use: "show <slug>", Short: "Show skill details", Args: cobra.ExactArgs(1), RunE: runSkillsShow},
		&cobra.Command{Use: "install <source>", Short: "Install a skill from GitHub", Args: cobra.ExactArgs(1), RunE: runSkillsInstall},
		&cobra.Command{Use: "delete <name>", Short: "Delete an installed skill", Args: cobra.ExactArgs(1), RunE: runSkillsDelete},
	)
	return cmd
}

func runSkillsList(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	sk := skills.NewSkillStore(filepath.Join(home, ".hi", "skills"))
	list, err := sk.List(context.Background())
	if err != nil {
		return err
	}
	for _, s := range list {
		always := ""
		if s.Frontmatter.AlwaysActive {
			always = " (always)"
		}
		fmt.Printf("%s  %s%s\n", s.Slug, s.Frontmatter.Description, always)
	}
	return nil
}

func runSkillsShow(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	sk := skills.NewSkillStore(filepath.Join(home, ".hi", "skills"))
	skill, err := sk.Get(context.Background(), args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Name: %s\n", skill.Frontmatter.Name)
	fmt.Printf("Slug: %s\n", skill.Slug)
	fmt.Printf("Description: %s\n", skill.Frontmatter.Description)
	fmt.Printf("Triggers: %v\n", skill.Frontmatter.Triggers)
	fmt.Printf("Always Active: %v\n", skill.Frontmatter.AlwaysActive)
	return nil
}

func runSkillsInstall(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	sk := skills.NewSkillStore(filepath.Join(home, ".hi", "skills"))
	skill, err := sk.Install(context.Background(), args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Installed %q (%s)\n", skill.Slug, skill.Frontmatter.Description)
	return nil
}

func runSkillsDelete(cmd *cobra.Command, args []string) error {
	home, _ := os.UserHomeDir()
	sk := skills.NewSkillStore(filepath.Join(home, ".hi", "skills"))
	return sk.Delete(context.Background(), args[0])
}
