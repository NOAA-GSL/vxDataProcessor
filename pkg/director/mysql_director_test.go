package director

import (
	"testing"
	"fmt"
	"os"
	"strings"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func Test_getMySqlConnection(t *testing.T) {

	type args struct {
		mysqlCredentials DbCredentials
	}
	var environmentFile string = fmt.Sprint(os.Getenv("HOME"), "/vxDataProcessor.env")
	err := godotenv.Load(environmentFile)
	if err != nil {
		t.Fatalf("Error loading .env file: %q", environmentFile)
	}
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
			name: "test_connection",
			args: args{mysqlCredentials: mysqlCredentials},
			want: "*sql.DB",
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

type stringScalar struct {
	avtime			int64
	squareDiffSum   float64
	NSum            int64
	obsModelDiffSum float64
	modelSum        float64
	obsSum          float64
	absSum          float64
}

//record.squareDiffSum record.NSum record.obsModelDiffSum record.modelSum record.obsSum record.absSum record.time
func Test_mySqlQuery(t *testing.T) {

	var environmentFile string = fmt.Sprint(os.Getenv("HOME"), "/vxDataProcessor.env")
	err := godotenv.Load(environmentFile)
	if err != nil {
		t.Fatalf("Error loading .env file: %q", environmentFile)
	}
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
	if (err != nil){
		t.Fatalf("getMySqlConnection() error = %v", err)
		return
	}
	defer mysqlDB.Close()
	tests := []struct {
		name    string
		args    string
		want    int
		wantErr bool
	}{
		{
			name: "test_query",
			args: "testdata/stmnt.sql",
			want: 667,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		buf, err := os.ReadFile(tt.args)
		if (err != nil){
			t.Fatalf("Test_mySqlQuery() error reading test statement= %v", err)
			return
		}
		stmnt := string(buf)
		fromEpoch := "1675281600" // Tue, 1 Feb 2023 20:00:00 GMT
		toEpoch := "1677700800"  // Tue, 1 Mar 2023 20:00:00 GMT
		stmnt = strings.ReplaceAll(stmnt, "{ { fromSecs } }", fromEpoch)
		stmnt = strings.ReplaceAll(stmnt, "{ { toSecs } }", toEpoch)
		t.Run(tt.name, func(t *testing.T) {
			var record stringScalar
			//square_diff_sum, &N_sum, &obs_model_diff_sum, &model_sum, &obs_sum, &abs_sum
			var records []stringScalar
			rows, err := mysqlDB.Query(stmnt)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMySqlConnection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			defer rows.Close()
			for rows.Next() {
				if err := rows.Scan(&record.avtime,&record.squareDiffSum, &record.NSum, &record.obsModelDiffSum, &record.modelSum, &record.obsSum, &record.absSum); err != nil {
					t.Errorf("could not scan row: %v", err)
				}
				records = append(records, record)
			}
			if tt.want != len(records){
				t.Errorf("Test_mySqlQuery() data length is wrong = %v", len(records))
			}
		})
	}
}
