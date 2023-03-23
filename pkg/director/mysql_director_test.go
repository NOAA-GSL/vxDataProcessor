package director

import (
	"database/sql"
	"reflect"
	"testing"
	_ "github.com/go-sql-driver/mysql"
)

func Test_getMySqlConnection(t *testing.T) {
	type args struct {
		mysqlCredentials DbCredentials
	}
	tests := []struct {
		name    string
		args    args
		want    *sql.DB
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMySqlConnection(tt.args.mysqlCredentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMySqlConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMySqlConnection() = %v, want %v", got, tt.want)
			}
		})
	}
}
