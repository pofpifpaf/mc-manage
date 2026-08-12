package config

import (
	"minecraft-manager/internal/paths"
	"os"
	"strings"
)

func NeoforgeConfigureJavaRunScript(server string) error {

	cfg, err := Load(server)
	if err != nil {
		return err
	}

	runPath := paths.Jar(server, cfg.Jar)
	javaPath := paths.Java(cfg.Java)

	data, err := os.ReadFile(runPath)
	if err != nil {
		return err
	}

	content := string(data)

	content = strings.Replace(content, "java @user_jvm_args.txt", javaPath+" @user_jvm_args.txt", 1)

	err = os.WriteFile(runPath, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func NeoforgeSetJVMArg(server string) error {
	_ = server // TODO: go into user_jvm_args.txt and add them there

	return nil
}
