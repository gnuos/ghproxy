package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	cfgPath string
	rootCmd = &cobra.Command{
		Use:     filepath.Base(os.Args[0]),
		Short:   "一个用于加速下载github和其他限速网站资源的代理服务",
		Version: VERSION,
		Run:     runServer,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "", "config file (default is ./cfg.hcl or $HOME/.cfg.hcl)")
}

func runServer(cmd *cobra.Command, args []string) {
	var err error

	if err = ensureConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if cfgPath == "" {
		fmt.Fprintf(os.Stderr, "Error: 没找到配置文件路径\n\n")

		cmd.Help()
		os.Exit(1)
	}

	cfg, err = ParseConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	startWeb()
}

func ensureConfig() (err error) {
	var cwd, userHome string

	cwd, err = os.Getwd()
	if err != nil {
		return
	}

	if cfgPath != "" {
		if fileReadable(cfgPath) {
			cfgPath, err = filepath.Abs(cfgPath)
			if err != nil {
				cfgPath = ""
				return
			}

			return
		}

		cfgPath = ""
	} else {
		cfgPath = cwd + "/cfg.hcl"
		if fileReadable(cfgPath) {
			return
		}

		userHome, err = os.UserHomeDir()
		if err != nil {
			cfgPath = ""
			return
		}

		cfgPath = userHome + "/.cfg.hcl"
		if dirExists(userHome) && fileReadable(cfgPath) {
			return
		}

		cfgPath = ""
	}

	return
}

func fileReadable(f string) bool {
	info, err := os.Stat(f)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	if info.Mode().Perm()&0444 != 0444 {
		return false
	}

	return true
}

func dirExists(d string) bool {
	info, err := os.Stat(d)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	if !info.IsDir() {
		return false
	}

	return true
}
