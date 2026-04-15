package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fanyang89/mirror-switch-cli/pkg/backup"
	"github.com/fanyang89/mirror-switch-cli/pkg/mirrors"
	"github.com/fanyang89/mirror-switch-cli/pkg/osdetect"
	"github.com/fanyang89/mirror-switch-cli/pkg/switcher"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	repoDir         string
	osReleasePath   string
	dryRun          bool
	interactive     bool
	disableOpenH264 bool
)

var rootCmd = &cobra.Command{
	Use:   "mirror-switch",
	Short: "Switch repository mirrors to CERNET mirror",
	Long:  "A CLI tool to automatically switch Linux package manager repository mirrors to the CERNET mirror.",
}

var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch repositories to CERNET mirror",
	Run: func(cmd *cobra.Command, args []string) {
		osdetect.OSReleasePath = osReleasePath
		info, err := osdetect.Detect()
		if err != nil {
			fmt.Printf("Error detecting OS: %v\n", err)
			return
		}

		fmt.Printf("Detected OS: %s (Version: %s)\n", info.ID, info.VersionID)

		availableRepoTypes := []switcher.RepoType{switcher.EPEL, switcher.RPMFusion}
		switch info.ID {
		case "rocky":
			availableRepoTypes = append(availableRepoTypes, switcher.Rocky)
		case "fedora":
			availableRepoTypes = append(availableRepoTypes, switcher.Fedora)
		}

		var selectedRepoTypes []switcher.RepoType
		if interactive {
			options := []string{}
			for _, rt := range availableRepoTypes {
				options = append(options, string(rt))
			}
			prompt := &survey.MultiSelect{
				Message: "Select repositories to switch:",
				Options: options,
				Default: options,
			}
			var selections []string
			if err := survey.AskOne(prompt, &selections); err != nil {
				fmt.Printf("Prompt failed: %v\n", err)
				return
			}
			for _, s := range selections {
				selectedRepoTypes = append(selectedRepoTypes, switcher.RepoType(s))
			}
		} else {
			selectedRepoTypes = availableRepoTypes
		}

		mirrorOptions := []string{}
		for _, m := range mirrors.MirrorNodes {
			mirrorOptions = append(mirrorOptions, fmt.Sprintf("%s [%s]", m.Name, m.Host))
		}

		for _, rt := range selectedRepoTypes {
			fmt.Printf("\nProcessing %s repositories...\n", rt)
			files, _ := filepath.Glob(filepath.Join(repoDir, fmt.Sprintf("%s*.repo", rt)))
			if len(files) == 0 {
				fmt.Printf("  No %s repo files found, skipping.\n", rt)
				continue
			}

			disableOpenH264ForRepo := disableOpenH264
			if interactive && rt == switcher.Fedora {
				openH264Files, _ := filepath.Glob(filepath.Join(repoDir, "fedora-cisco-openh264*.repo"))
				if len(openH264Files) > 0 {
					defaultAction := "Skip"
					if disableOpenH264ForRepo {
						defaultAction = "Disable"
					}

					var openH264Action string
					prompt := &survey.Select{
						Message: "fedora-cisco-openh264 repositories detected:",
						Options: []string{"Skip", "Disable"},
						Default: defaultAction,
					}
					if err := survey.AskOne(prompt, &openH264Action); err != nil {
						fmt.Printf("  Prompt failed: %v\n", err)
						continue
					}

					disableOpenH264ForRepo = openH264Action == "Disable"
					fmt.Printf("  fedora-cisco-openh264: %s\n", strings.ToLower(openH264Action))
				}
			}

			selectedMirrorHost := "mirrors.cernet.edu.cn"
			if interactive {
				var mirrorSelection string
				prompt := &survey.Select{
					Message: fmt.Sprintf("Select a mirror site for %s:", rt),
					Options: mirrorOptions,
					Default: mirrorOptions[0],
				}
				if err := survey.AskOne(prompt, &mirrorSelection); err != nil {
					fmt.Printf("  Prompt failed: %v\n", err)
					continue
				}
				// Extract host from "Name [host]"
				start := strings.LastIndex(mirrorSelection, "[")
				end := strings.LastIndex(mirrorSelection, "]")
				if start != -1 && end != -1 {
					selectedMirrorHost = mirrorSelection[start+1 : end]
				}
			}
			fmt.Printf("  Using mirror site: %s\n", selectedMirrorHost)

			if !dryRun {
				for _, f := range files {
					_, err := backup.CreateBackup(f)
					if err != nil {
						fmt.Printf("  Warning: Failed to create backup for %s: %v\n", f, err)
					}
				}

				switched, err := switcher.Switch(rt, switcher.Config{
					RepoDir:         repoDir,
					MirrorHost:      selectedMirrorHost,
					DisableOpenH264: disableOpenH264ForRepo,
				})
				if err != nil {
					fmt.Printf("  Error switching %s: %v\n", rt, err)
				} else {
					fmt.Printf("  Successfully switched %d files for %s.\n", len(switched), rt)
				}
			} else {
				fmt.Printf("  (Dry run) Would switch %d files for %s using %s.\n", len(files), rt, selectedMirrorHost)
			}
		}
	},
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore repository configurations from backup",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Restoring backups in %s...\n", repoDir)
		restored, err := backup.RestoreBackups(repoDir)
		if err != nil {
			fmt.Printf("Error restoring backups: %v\n", err)
			return
		}
		fmt.Printf("Successfully restored %d repository files.\n", len(restored))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&repoDir, "repo-dir", "d", "/etc/yum.repos.d/", "Directory containing repository files")
	rootCmd.PersistentFlags().StringVar(&osReleasePath, "os-release", "/etc/os-release", "Path to os-release file")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "i", false, "Interactive mode")
	switchCmd.Flags().BoolVar(&disableOpenH264, "disable-openh264", false, "Disable fedora-cisco-openh264 repositories instead of switching them")

	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(restoreCmd)
}
