package cli

import (
	"github.com/breinzhang/tokusage/internal/app"
	"github.com/breinzhang/tokusage/internal/platform"
	"github.com/spf13/cobra"
)

func newCacheCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage tokusage local cache",
	}
	cmd.AddCommand(newCacheStatusCommand())
	cmd.AddCommand(newCacheRebuildCommand())
	cmd.AddCommand(newCacheClearCommand())
	return cmd
}

func newCacheStatusCommand() *cobra.Command {
	var cachePath string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cache status",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cachePathOrDefault(cachePath)
			if err != nil {
				return err
			}
			out, err := app.RunCacheStatus(cmd.Context(), path)
			if err != nil {
				return err
			}
			cmd.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	return cmd
}

func newCacheRebuildCommand() *cobra.Command {
	var cachePath string
	cmd := &cobra.Command{
		Use:   "rebuild",
		Short: "Rebuild usage cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cachePathOrDefault(cachePath)
			if err != nil {
				return err
			}
			opts, err := claudeOptions(nil, path, "", "", "", "table", "day", false)
			if err != nil {
				return err
			}
			out, err := app.RunCacheRebuild(cmd.Context(), opts)
			if err != nil {
				return err
			}
			cmd.Print(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	return cmd
}

func newCacheClearCommand() *cobra.Command {
	var cachePath string
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear usage cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := cachePathOrDefault(cachePath)
			if err != nil {
				return err
			}
			return app.RunCacheClear(path)
		},
	}
	cmd.Flags().StringVar(&cachePath, "cache", "", "cache database path")
	return cmd
}

func cachePathOrDefault(cachePath string) (string, error) {
	if cachePath != "" {
		return cachePath, nil
	}
	return platform.DefaultCacheDBPath()
}
