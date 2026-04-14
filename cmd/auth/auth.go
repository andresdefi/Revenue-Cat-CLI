package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	internalAuth "github.com/andresdefi/rc/internal/auth"
	"github.com/andresdefi/rc/internal/cmdutil"
	"github.com/andresdefi/rc/internal/output"
	"github.com/spf13/cobra"
)

func NewAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long:  "Log in, check status, or log out of the RevenueCat API.",
	}

	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newStatusCmd())
	cmd.AddCommand(newLogoutCmd())
	return cmd
}

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with a RevenueCat API v2 secret key",
		Long: `Authenticate with your RevenueCat API v2 secret key.

You can create a v2 secret key in the RevenueCat dashboard:
  Project Settings > API Keys > + New Secret API Key

The key will be stored in your system keychain (with config file fallback).
Keys are prefixed with sk_ and must have v2 API permissions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			fmt.Print("Enter your RevenueCat API v2 secret key: ")
			reader := bufio.NewReader(os.Stdin)
			token, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			token = strings.TrimSpace(token)

			if token == "" {
				return fmt.Errorf("API key cannot be empty")
			}

			if !strings.HasPrefix(token, "sk_") {
				output.Warn("Key does not start with 'sk_' - make sure you're using a v2 secret API key")
			}

			if err := internalAuth.SaveToken(profile, token); err != nil {
				return fmt.Errorf("failed to save API key: %w", err)
			}

			output.Success("Logged in successfully [profile: %s] (stored in %s)", profile, internalAuth.TokenSource(profile))
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			token, err := internalAuth.GetToken(profile)
			if err != nil {
				fmt.Printf("Not logged in [profile: %s]\n", profile)
				fmt.Println("Run `rc auth login` to authenticate")
				return nil
			}

			fmt.Printf("Logged in [profile: %s]\n", profile)
			fmt.Printf("  Key:     %s\n", internalAuth.MaskToken(token))
			fmt.Printf("  Stored:  %s\n", internalAuth.TokenSource(profile))
			return nil
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove stored API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := cmdutil.ResolveProfile()

			if err := internalAuth.DeleteToken(profile); err != nil {
				return fmt.Errorf("failed to remove API key: %w", err)
			}
			output.Success("Logged out successfully [profile: %s]", profile)
			return nil
		},
	}
}
