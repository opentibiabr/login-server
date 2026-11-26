package configs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPvpTypeFromWorldType(t *testing.T) {
	tests := []struct {
		worldType string
		want      uint32
		known     bool
	}{
		{"retro-pvp", PvpTypeRetroOpen, true},
		{"RETRO-PVP", PvpTypeRetroOpen, true},
		{" retro-pvp ", PvpTypeRetroOpen, true},
		{"pvp", PvpTypeRetroOpen, true},
		{"retro-pvp-enforced", PvpTypeRetroHardcore, true},
		{"no-pvp", PvpTypeOptional, true},
		{"pvp-enforced", PvpTypeHardcore, true},
		{"expert-pvp", PvpTypeOpen, true},
		{"retro", PvpTypeOpen, false},
		{"", PvpTypeOpen, false},
	}
	for _, tt := range tests {
		got, known := PvpTypeFromWorldType(tt.worldType)
		assert.Equal(t, tt.want, got, "worldType %q", tt.worldType)
		assert.Equal(t, tt.known, known, "worldType %q", tt.worldType)
	}
}

func TestPvpTypeCodesMatchTheClientList(t *testing.T) {
	// Wire values read by the client: changing one of them changes the label players see for the world.
	assert.Equal(t, uint32(0), PvpTypeOpen)
	assert.Equal(t, uint32(1), PvpTypeOptional)
	assert.Equal(t, uint32(2), PvpTypeHardcore)
	assert.Equal(t, uint32(3), PvpTypeRetroOpen)
	assert.Equal(t, uint32(4), PvpTypeRetroHardcore)
}

func TestGetServerPvpType(t *testing.T) {
	serverWith := func(t *testing.T, config string) string {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.lua"), []byte(config), 0o600))
		return dir
	}
	retroConfig := "serverName = \"Test\"\n-- valid values for worldType are ...\nworldType = \"retro-pvp\" -- comment\n"

	tests := []struct {
		name      string
		envValue  string
		serverDir func(t *testing.T) string
		want      uint32
	}{
		{name: "nothing configured keeps open PvP", want: PvpTypeOpen},
		{name: "a world type name in the variable", envValue: "retro-pvp", want: PvpTypeRetroOpen},
		{name: "a code in the variable", envValue: "3", want: PvpTypeRetroOpen},
		{name: "code zero in the variable", envValue: "0", want: PvpTypeOpen},
		{name: "a code above four is unknown", envValue: "5", want: PvpTypeOpen},
		{name: "an unknown name falls back to open PvP", envValue: "retro", want: PvpTypeOpen},
		{name: "worldType of the server configuration", serverDir: func(t *testing.T) string { return serverWith(t, retroConfig) }, want: PvpTypeRetroOpen},
		{
			name:      "the variable wins over the server configuration",
			envValue:  "no-pvp",
			serverDir: func(t *testing.T) string { return serverWith(t, retroConfig) },
			want:      PvpTypeOptional,
		},
		{name: "a server path without a configuration", serverDir: func(t *testing.T) string { return t.TempDir() }, want: PvpTypeOpen},
		{name: "a configuration without worldType", serverDir: func(t *testing.T) string { return serverWith(t, "serverName = \"Test\"\n") }, want: PvpTypeOpen},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(EnvServerPvpTypeKey, tt.envValue)
			serverPath := ""
			if tt.serverDir != nil {
				serverPath = tt.serverDir(t)
			}
			t.Setenv(EnvServerPathKey, serverPath)
			assert.Equal(t, tt.want, GetServerPvpType())
		})
	}
}
