package users

import (
	"errors"
	"fmt"
	"minecraft-manager/internal/paths"
	"minecraft-manager/internal/protocol"
	"minecraft-manager/internal/ui"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
)

func CreateUser(cfg *protocol.Config) error {

	cfg.Username = "mc-" + cfg.Name

	cmd := exec.Command(
		"useradd",
		"--system",
		"--no-create-home",
		"--shell", "/usr/sbin/nologin",
		"--user-group",
		cfg.Username,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("useradd failed: %w: %s", err, output)
	}

	u, err := user.Lookup(cfg.Username)
	if err != nil {
		return err
	}

	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return err
	}

	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		return err
	}

	cfg.Uid = int(uid)
	cfg.Gid = int(gid)

	ui.PrintSuccess(fmt.Sprintf("Created user %s with uid: %d and gid: %d", cfg.Username, cfg.Uid, cfg.Gid))

	return nil
}

func createUserGivenUIDGID(cfg *protocol.Config) error {

	ui.PrintInfo(fmt.Sprintf("Creating user %s with uid %d, gid %d", cfg.Username, cfg.Uid, cfg.Gid))

	uidStr := strconv.Itoa(cfg.Uid)
	gidStr := strconv.Itoa(cfg.Gid)

	cmd := exec.Command(
		"useradd",
		"--system",
		"--no-create-home",
		"--shell", "/usr/sbin/nologin",
		"-u", uidStr,
		"--user-group",
		cfg.Username,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("useradd failed: %w: %s", err, output)
	}

	u, err := user.Lookup(cfg.Username)
	if err != nil {
		return err
	}

	if u.Uid == uidStr && u.Gid == gidStr {
		return nil
	} else {
		return fmt.Errorf("Unknown error")
	}
}

func RemoveUser(cfg *protocol.Config) error {

	ui.PrintInfo(fmt.Sprintf("Removing user %s with uid %d, gid %d", cfg.Username, cfg.Uid, cfg.Gid))

	cmd := exec.Command(
		"userdel",
		cfg.Username,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("userdel failed: %w: %s", err, output)
	}

	cfg.Username = "disabled"
	cfg.Uid = -1
	cfg.Gid = -1

	return nil
}

func setFolderPermissions(directory string, uid, gid int) error {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	if err := os.Chown(directory, uid, gid); err != nil {
		return err
	}

	if err := os.Chmod(directory, 0700); err != nil {
		return err
	}

	for _, entry := range entries {

		entryPathAbsolute := filepath.Join(directory, entry.Name())

		if entry.IsDir() {
			if err := setFolderPermissions(entryPathAbsolute, uid, gid); err != nil {
				ui.PrintWarning("couldn't set permissions for folder: " + err.Error())
			}
			continue
		}

		if err := os.Chown(entryPathAbsolute, uid, gid); err != nil {
			ui.PrintWarning("couldn't own file: " + err.Error())
		}

		if err := os.Chmod(entryPathAbsolute, 0700); err != nil {
			ui.PrintWarning("couldn't set permissions: " + err.Error())
		}

	}

	return nil
}

func SetServerPermissions(cfg *protocol.Config) error {

	serverDir := paths.Server(cfg.Name)
	uid := cfg.Uid
	gid := cfg.Gid

	ui.PrintInfo("Setting folder permissions for user")

	return setFolderPermissions(serverDir, uid, gid)
}

func SetJarPermissions(cfg *protocol.Config) {

	ui.PrintInfo("Setting jar permissions")

	serverJarPath := paths.Jar(cfg.Name, cfg.Jar)
	uid := cfg.Uid
	gid := cfg.Gid

	if err := os.Chown(serverJarPath, uid, gid); err != nil {
		ui.PrintWarning("couldn't own file: " + err.Error())
	}

	if err := os.Chmod(serverJarPath, 0700); err != nil {
		ui.PrintWarning("couldn't set permissions: " + err.Error())
	}
}

func EnsureUserExistence(cfg *protocol.Config) error {

	u, err := user.Lookup(cfg.Username)
	_, ok := errors.AsType[user.UnknownUserError](err)
	if err == nil && !ok {

		uid, err := strconv.ParseUint(u.Uid, 10, 32)
		if err != nil {
			return err
		}

		gid, err := strconv.ParseUint(u.Gid, 10, 32)
		if err != nil {
			return err
		}

		if cfg.Uid == int(uid) && cfg.Gid == int(gid) {
			return nil
		} else {
			return fmt.Errorf("User already exists, with different uid/gid")
		}
	}

	return createUserGivenUIDGID(cfg)
}

func EnsureUserExistenceServerInfo(server protocol.ServerInfo) error {
	var cfg protocol.Config

	cfg.Username = server.Username
	cfg.Uid = server.Uid
	cfg.Gid = server.Gid

	return EnsureUserExistence(&cfg)
}
