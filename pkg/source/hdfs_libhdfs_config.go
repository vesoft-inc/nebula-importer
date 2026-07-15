package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var libHDFSConfigDirCache sync.Map

func prepareLibHDFSEnv(c *HDFSConfig) error {
	if c.HadoopHome != "" {
		if err := ensureLibHDFSEnv("HADOOP_HOME", c.HadoopHome); err != nil {
			return err
		}
	}

	if c.JavaHome != "" {
		if err := ensureLibHDFSEnv("JAVA_HOME", c.JavaHome); err != nil {
			return err
		}
	}

	confDir, err := libHDFSConfigDir(c)
	if err != nil {
		return err
	}
	if confDir != "" {
		if err = ensureLibHDFSEnv("HADOOP_CONF_DIR", confDir); err != nil {
			return err
		}
	}

	if c.Krb5ConfigFile != "" {
		if err = ensureLibHDFSEnv("KRB5_CONFIG", c.Krb5ConfigFile); err != nil {
			return err
		}
	}

	return nil
}

func ensureLibHDFSEnv(key, value string) error {
	currentValue, found := os.LookupEnv(key)
	switch {
	case !found || currentValue == "":
		return os.Setenv(key, value)
	case currentValue == value:
		return nil
	default:
		return fmt.Errorf("%s is already set to %q, can not override it with %q for the libhdfs backend", key, currentValue, value)
	}
}

func libHDFSConfigDir(c *HDFSConfig) (string, error) {
	if c.HadoopConfigDir != "" {
		return c.HadoopConfigDir, nil
	}

	if c.CoreSiteFile == "" && c.HDFSSiteFile == "" {
		return "", nil
	}

	cacheKey := c.CoreSiteFile + "\x00" + c.HDFSSiteFile
	if cachedDir, ok := libHDFSConfigDirCache.Load(cacheKey); ok {
		return cachedDir.(string), nil
	}

	dir, err := os.MkdirTemp("", "nebula-importer-hadoop-conf-*")
	if err != nil {
		return "", err
	}

	if err = installLibHDFSConfigFile(c.CoreSiteFile, dir, "core-site.xml"); err != nil {
		return "", err
	}
	if err = installLibHDFSConfigFile(c.HDFSSiteFile, dir, "hdfs-site.xml"); err != nil {
		return "", err
	}
	libHDFSConfigDirCache.Store(cacheKey, dir)
	return dir, nil
}

func installLibHDFSConfigFile(src, dir, name string) error {
	if src == "" {
		return nil
	}

	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(dir, name)
	return os.WriteFile(targetPath, content, 0o644)
}
