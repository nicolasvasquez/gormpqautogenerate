package src

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

const queryTables = "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE'"
const queryFields = "SELECT column_name, data_type, is_nullable, column_default FROM information_schema.columns WHERE table_name ='"

var fieldTypes map[string]string

type model struct {
	Name   string
	Fields []field
}

type field struct {
	Name     string
	Type     string
	Nullable string
	Default  string
}

//Generate models from database scheme. Only work in postgres
func Generate() {
	setDataType()
	db := getDatabase()
	columns := getTableModels(db)
	for _, el := range columns {
		createFile(el)
	}
}

func setDataType() {
	fieldTypes = map[string]string{
		"bigint":                   "int64",
		"integer":                  "int",
		"smallint":                 "int",
		"double precision":         "float64",
		"character varying":        "string",
		"character":                "string",
		"text":                     "string",
		"bytea":                    "[]byte",
		"date":                     "time.Time",
		"datetime":                 "time.Time",
		"timestamp":                "time.Time",
		"timestamp with time zone": "time.Time",
		"numeric":                  "float64",
		"decimal":                  "float64",
		"bit":                      "uint64",
		"boolean":                  "bool",
	}
}

func getDatabaseConnectionString(name, pass, host, port, database string) string {
	return "postgres://" +
		name +
		":" +
		pass +
		"@" +
		host +
		":" +
		port +
		"/" +
		database +
		"?sslmode=disable"
}

func getDatabase() *sql.DB {
	fmt.Printf(getDatabaseConnectionString("test", "test", "localhost", "5432", "test") + "\n")
	db, err := sql.Open("postgres", getDatabaseConnectionString("test", "test", "localhost", "5432", "test"))
	if err != nil {
		panic(err)
	}
	return db
}

func getTableModels(db *sql.DB) []model {
	var tables []model
	qry, err := db.Query(queryTables)
	if err != nil {
		panic(err)
	}
	for qry.Next() {
		table := model{}
		qry.Scan(&table.Name)
		table.Fields = setTableFields(db, table.Name)
		tables = append(tables, table)
	}
	return tables
}

func setTableFields(db *sql.DB, tableName string) []field {
	var fields []field
	qry, err := db.Query(queryFields + tableName + "'")
	if err != nil {
		panic(err)
	}
	for qry.Next() {
		column := field{}
		qry.Scan(&column.Name, &column.Type, &column.Nullable, &column.Default)
		fields = append(fields, column)
	}
	return fields
}

func createFile(table model) {
	f, err := os.Create("out/models/" + table.Name + ".go")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	addHeader(f)
	addStruct(f, table)
}

func addHeader(f *os.File) {
	f.WriteString("package models\n\n")
}

func addStruct(f *os.File, table model) {
	tableName := strings.Replace(strings.Title(strings.Replace(table.Name, "_", " ", -1)), " ", "", -1)
	f.WriteString("type " + tableName + " struct {\n")
	fmt.Printf("Name=%s\n", table.Name)
	for _, el := range table.Fields {
		fmt.Printf("Column=%s, Type=%s, Null=%s, DefaultValue=%s\n", el.Name, el.Type, el.Nullable, el.Default)
		f.WriteString(parseField(el, table.Name))
		f.WriteString("\n")
	}
	f.WriteString("}")
}

func parseField(column field, modelName string) string {
	name := strings.Replace(strings.Title(strings.Replace(column.Name, "_", " ", -1)), " ", "", -1)
	tp := fieldTypes[column.Type]
	sql := "`gorm:\"column:" + column.Name + ";"
	if column.Nullable == "NO" {
		sql += "not null;"
	}
	sql += `"`
	sql += " json:\"" + column.Name + "\""
	sql += " form:\"" + modelName + "_" + column.Name + "\"`"
	line := fmt.Sprintf("  %-10s\t%-10s\t%-20s", name, tp, sql)
	return line
}
