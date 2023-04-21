//go:build integration
// +build integration

package director

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/NOAA-GSL/vxDataProcessor/pkg/builder"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func loadEnvironmentFile() {
	environmentFile, set := os.LookupEnv("PROC_ENV_PATH")
	if !set {
		err := godotenv.Load() // Loads from "$(pwd)/.env"
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
		}
	} else {
		err := godotenv.Load(environmentFile) // Loads from whatever PROC_ENV_PATH has been set to
		if err != nil {
			log.Printf("Couldn't load environment file: %q", environmentFile)
		}
	}
}

func Test_getMySqlConnection(t *testing.T) {
	type args struct {
		mysqlCredentials DbCredentials
	}
	loadEnvironmentFile()
	var mysqlCredentials DbCredentials
	// refer to https://github.com/go-sql-driver/mysql/#dsn-data-source-name
	mysqlCredentials.Host = os.Getenv("MYSQL_HOST")
	if mysqlCredentials.Host == "" {
		t.Fatalf("Undefined MYSQL_HOST in environment")
	}
	mysqlCredentials.User = os.Getenv("MYSQL_USER")
	if mysqlCredentials.User == "" {
		t.Fatalf("Undefined MYSQL_USER in environment")
	}
	mysqlCredentials.Password = os.Getenv("MYSQL_PASSWORD")
	if mysqlCredentials.Password == "" {
		t.Fatalf("Undefined MYSQL_PASSWORD in environment")
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "test_connection",
			args:    args{mysqlCredentials: mysqlCredentials},
			want:    "*sql.DB",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMySqlConnection(tt.args.mysqlCredentials)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMySqlConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer got.Close()
			if tt.want != fmt.Sprintf("%T", got) {
				t.Errorf(fmt.Sprintf("getMySqlConnection() type of connection is not sql.DB = %v", fmt.Sprintf("%T", got)))
			}
		})
	}
}

// record.squareDiffSum record.NSum record.obsModelDiffSum record.modelSum record.obsSum record.absSum record.time
func Test_mySqlQuery(t *testing.T) {
	loadEnvironmentFile()
	var mysqlCredentials DbCredentials
	// refer to https://github.com/go-sql-driver/mysql/#dsn-data-source-name
	mysqlCredentials.Host = os.Getenv("MYSQL_HOST")
	if mysqlCredentials.Host == "" {
		t.Fatalf("Undefined MYSQL_HOST in environment")
	}
	mysqlCredentials.User = os.Getenv("MYSQL_USER")
	if mysqlCredentials.User == "" {
		t.Fatalf("Undefined MYSQL_USER in environment")
	}
	mysqlCredentials.Password = os.Getenv("MYSQL_PASSWORD")
	if mysqlCredentials.Password == "" {
		t.Fatalf("Undefined MYSQL_PASSWORD in environment")
	}
	mysqlDB, err := getMySqlConnection(mysqlCredentials)
	if err != nil {
		t.Fatalf("getMySqlConnection() error = %v", err)
		return
	}
	defer mysqlDB.Close()
	tests := []struct {
		name       string
		args       string
		fromEpoch  string
		toEpoch    string
		recordType string
		want       int
		wantErr    bool
	}{
		{
			name:       "test_query_scalar",
			args:       "testdata/scalar_stmnt.sql",
			recordType: "scalar",
			fromEpoch:  "1675281600", // Tue, 1 Feb 2023 20:00:00 GMT
			toEpoch:    "1677700800", // Tue, 1 Mar 2023 20:00:00 GMT
			want:       667,
			wantErr:    false,
		},
		{
			name:       "test_query_ctc",
			args:       "testdata/ctc_stmnt.sql",
			recordType: "ctc",
			fromEpoch:  "1675281600", // Tue, 1 Feb 2023 20:00:00 GMT
			toEpoch:    "1677700800", // Tue, 1 Mar 2023 20:00:00 GMT
			want:       613,
			wantErr:    false,
		},
		{
			name:       "test_query_precalc",
			args:       "testdata/precalc_stmnt.sql",
			recordType: "precalc",
			fromEpoch:  "1587513600", // Wednesday, April 22, 2020 12:00:00 AM
			toEpoch:    "1631620800", // Tuesday, September 14, 2021 12:00:00 PM
			want:       1000,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		buf, err := os.ReadFile(tt.args)
		if err != nil {
			t.Fatalf("Test_mySqlQuery() error reading test statement= %v", err)
			return
		}

		stmnt := string(buf)
		fromEpoch := tt.fromEpoch // Tue, 1 Feb 2023 20:00:00 GMT
		toEpoch := tt.toEpoch     // Tue, 1 Mar 2023 20:00:00 GMT
		stmnt = strings.ReplaceAll(stmnt, "{ { fromSecs } }", fromEpoch)
		stmnt = strings.ReplaceAll(stmnt, "{ { toSecs } }", toEpoch)
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			rows, err := mysqlDB.Query(stmnt)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMySqlConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var records []interface{}
			defer rows.Close()
			for rows.Next() {
				switch tt.recordType {
				case "scalar":
					var record builder.ScalarRecord
					if err := rows.Scan(&record.Avtime, &record.SquareDiffSum, &record.NSum, &record.ObsModelDiffSum, &record.ModelSum, &record.ObsSum, &record.AbsSum); err != nil {
						t.Errorf("could not scan row: %v", err)
					} else {
						records = append(records, record)
					}
				case "ctc":
					var record builder.CTCRecord
					if err := rows.Scan(&record.Avtime, &record.Hit, &record.Miss, &record.Fa, &record.Cn); err != nil {
						t.Errorf("could not scan row: %v", err)
					}
					records = append(records, record)
				case "precalc":
					var record builder.PreCalcRecord
					if err := rows.Scan(&record.Avtime, &record.Stat); err != nil {
						t.Errorf("could not scan row: %v", err)
					}
					records = append(records, record)
				default:
					t.Fatalf("Test_mySqlQuery unrecognized record type %q", tt.recordType)
				}
			}
			elapsed := time.Since(start)
			fmt.Printf("The query and scan took combined %s", elapsed)
			if tt.want != len(records) {
				t.Errorf("Test_mySqlQuery() data length is wrong = %v", len(records))
			}
		})
	}
}
