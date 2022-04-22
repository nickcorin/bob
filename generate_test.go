package bob_test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nickcorin/adventech/bob"
)

var update = flag.Bool("update", false, "Update the golden files")

func TestGenerate(t *testing.T) {
	conf := bob.Config{
		Docker: &bob.DockerConfig{
			RestartPolicy: "always",
		},
		Go: &bob.GoConfig{
			Arch:      "amd64",
			BinaryDir: "bin",
			Module:    "github.com/nickcorin/adventech",
			OS:        "linux",
		},
		Service: &bob.ServiceConfig{
			Name:        "test",
			Host:        "test.example.com",
			Ports:       []string{"12345:12345"},
			MetricsPort: 9090,
		},
		Requires: []string{"Dockerfile", "Makefile", "ServiceDiscovery"},
	}

	configData, err := json.Marshal(&conf)
	require.NoError(t, err)

	serviceDir := filepath.Join("/tmp", conf.Service.Name)
	err = os.Mkdir(serviceDir, 0o755)
	require.NoError(t, err)

	metricsDir := filepath.Join("/tmp", "prometheus")
	err = os.Mkdir(metricsDir, 0o755)
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(serviceDir, "bob.json"))
	require.NoError(t, err)

	err = ioutil.WriteFile(tmpFile.Name(), configData, 0o755)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = os.RemoveAll(serviceDir)
		require.NoError(t, err)

		err = os.RemoveAll(metricsDir)
		require.NoError(t, err)
	})

	err = bob.GenerateService(tmpFile.Name())
	require.NoError(t, err)

	t.Run("generate dockerfile", func(t *testing.T) {
		actual, err := ioutil.ReadFile(filepath.Join(serviceDir, "Dockerfile"))
		require.NoError(t, err)
		require.NotNil(t, actual)

		goldenPath := filepath.Join("testdata/Dockerfile.golden")
		assertGolden(t, goldenPath, actual)
	})

	t.Run("generate makefile", func(t *testing.T) {
		actual, err := ioutil.ReadFile(filepath.Join(serviceDir, "Makefile"))
		require.NoError(t, err)
		require.NotNil(t, actual)

		goldenPath := filepath.Join("testdata/Makefile.golden")
		assertGolden(t, goldenPath, actual)
	})

	t.Run("generate service discovery", func(t *testing.T) {
		actual, err := ioutil.ReadFile(filepath.Join(metricsDir, "test.sd.json"))
		require.NoError(t, err)
		require.NotNil(t, actual)

		goldenPath := filepath.Join("testdata/ServiceDiscovery.golden")
		assertGolden(t, goldenPath, actual)
	})
}

func assertGolden(t *testing.T, goldenFile string, actual []byte) {
	t.Helper()

	if *update {
		err := ioutil.WriteFile(goldenFile, actual, 0o644)
		require.NoError(t, err)
		return
	}

	_, err := os.Stat(goldenFile)
	require.NoError(t, err)

	expected, err := ioutil.ReadFile(goldenFile)
	require.NoError(t, err)
	require.NotNil(t, actual)

	require.Equal(t, string(expected), string(actual))
}
