package configs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDBConfigs_Format(t *testing.T) {
	type fields struct {
		Host   string
		Port   int
		Name   string
		User   string
		Pass   string
		Config Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Format database properly",
		fields: fields{
			Host: "host",
			Port: 3510,
			Name: "mydb",
		},
		want: "Database: host:3510/mydb",
	}, {
		name: "Format database ignores user and pass",
		fields: fields{
			Host: "host",
			Port: 3510,
			Name: "mydb",
			User: "user",
			Pass: "pass",
		},
		want: "Database: host:3510/mydb",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConfigs := &DBConfigs{
				Host:   tt.fields.Host,
				Port:   tt.fields.Port,
				Name:   tt.fields.Name,
				User:   tt.fields.User,
				Pass:   tt.fields.Pass,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, dbConfigs.format())
		})
	}
}

func TestDBConfigs_GetConnectionString(t *testing.T) {
	type fields struct {
		Host   string
		Port   int
		Name   string
		User   string
		Pass   string
		Config Config
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{{
		name: "Build database connection string",
		fields: fields{
			Host: "host",
			Port: 3510,
			Name: "mydb",
			User: "user",
			Pass: "pass",
		},
		want: "user:pass@tcp(host:3510)/mydb?parseTime=true",
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConfigs := &DBConfigs{
				Host:   tt.fields.Host,
				Port:   tt.fields.Port,
				Name:   tt.fields.Name,
				User:   tt.fields.User,
				Pass:   tt.fields.Pass,
				Config: tt.fields.Config,
			}
			assert.Equal(t, tt.want, dbConfigs.GetConnectionString())
		})
	}
}

func TestGetDBConfigs(t *testing.T) {
	tests := []struct {
		name string
		want DBConfigs
	}{{
		name: "Default DB Configs",
		want: DBConfigs{
			Host: "127.0.0.1",
			Port: 3306,
			Name: "canary",
			User: "canary",
			Pass: "canary",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GetDBConfigs())
		})
	}
}
