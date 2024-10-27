package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gabe565.com/castsponsorskip/internal/config/names"
	"gabe565.com/castsponsorskip/internal/config/sponsorblockcast"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/structs"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/cobra"
	castdns "github.com/vishen/go-chromecast/dns"
)

const EnvPrefix = "CSS_"

var ErrInvalidIP = errors.New("failed to parse IP")

func Load(cmd *cobra.Command) (*Config, error) {
	k := koanf.New(".")
	c := New()

	// Load default config
	if err := k.Load(structs.Provider(c, "yaml"), nil); err != nil {
		return nil, err
	}

	// Find config file
	cfgFiles := make([]string, 0, 4)
	var fileRequired bool
	if cfgFile, err := cmd.Flags().GetString(names.FlagConfig); err != nil {
		panic(err)
	} else if cfgFile != "" {
		cfgFiles = append(cfgFiles, cfgFile)
		fileRequired = true
	} else {
		var configDir string
		if xdgConfigDir, err := os.UserConfigDir(); err == nil {
			configDir = xdgConfigDir
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}

			configDir = filepath.Join(home, ".config")
		}

		cfgFiles = append(cfgFiles,
			filepath.Join(configDir, "sponsorblockcast", "config.yaml"),
			filepath.Join(configDir, "sponsorblockcast", "config.yml"),
			filepath.Join("etc", "sponsorblockcast", "config.yaml"),
			filepath.Join("etc", "sponsorblockcast", "config.yml"),
		)
	}

	// Load config file
	parser := yaml.Parser()
	for _, cfgFile := range cfgFiles {
		if err := k.Load(file.Provider(cfgFile), parser); err != nil {
			if !fileRequired && errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		break
	}

	// Load deprecated envs
	if err := k.Load(sponsorblockcast.Provider(), nil); err != nil {
		return nil, err
	}

	// Load envs
	if err := k.Load(env.ProviderWithValue(EnvPrefix, ".", func(k string, v string) (string, any) {
		k = strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(k, EnvPrefix)), "_", "-")

		switch k {
		case names.FlagDevices, names.FlagCategories, names.FlagActionTypes:
			return k, strings.Split(v, ",")
		default:
			return k, v
		}
	}), nil); err != nil {
		return nil, err
	}

	// Load flags
	if err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil); err != nil {
		return nil, err
	}

	if err := k.UnmarshalWithConf("", &c, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return nil, err
	}

	c.InitLog(cmd.ErrOrStderr())

	if c.NetworkInterfaceName != "" {
		var err error
		if c.NetworkInterface, err = net.InterfaceByName(c.NetworkInterfaceName); err != nil {
			return nil, err
		}
	}

	for i, category := range c.Categories {
		c.Categories[i] = strings.TrimSpace(category)
	}

	for i, actionType := range c.ActionTypes {
		c.ActionTypes[i] = strings.TrimSpace(actionType)
	}

	if len(c.DeviceAddrStrs) != 0 {
		c.DeviceAddrs = make([]castdns.CastEntry, 0, len(c.DeviceAddrStrs))
		for _, device := range c.DeviceAddrStrs {
			u := url.URL{Host: device}

			castEntry := castdns.CastEntry{
				DeviceName: device,
				UUID:       device,
			}

			if port := u.Port(); port == "" {
				castEntry.Port = 8009
			} else {
				port, err := strconv.ParseUint(port, 10, 16)
				if err != nil {
					return nil, err
				}

				castEntry.Port = int(port) //nolint:gosec
			}

			if ip := net.ParseIP(u.Hostname()); ip == nil {
				return nil, fmt.Errorf("%w: %q", ErrInvalidIP, device)
			} else if ip.To4() != nil {
				castEntry.AddrV4 = ip
			} else {
				castEntry.AddrV6 = ip
			}

			c.DeviceAddrs = append(c.DeviceAddrs, castEntry)
		}
	}

	return c, nil
}
