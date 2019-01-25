package database

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"bitballot/config"
)

const (
	SQL_PROFILE = "select concat(title, firstname, lastname) from profiles where id $1 limit 1"
)

func Init(pathInit string) []string {
	return createTable(pathInit)
}

type Fields struct {
	ID uint64 `sql:"pk"`

	Createdby, Updatedby   uint64
	Createdate, Updatedate time.Time

	Title    string `sql:"index"`
	Workflow string `sql:"index"`
}

func (fi Fields) sqlCreate(table Tables) string {
	reflectType := reflect.TypeOf(table).Elem()
	tablename := strings.ToLower(reflectType.Name())
	sqlDrop := "drop table " + tablename

	sqlIndex := ""
	sqlCreate := "create table " + tablename + " ("
	sqlCreate, sqlIndex = sqlCreateFields(reflectType, tablename, sqlCreate, sqlIndex)
	sqlCreate = strings.TrimSuffix(sqlCreate, ", ") + "); "

	config.Get().Postgres.Exec(sqlDrop)
	_, err := config.Get().Postgres.Exec(sqlCreate)
	if err != nil {
		log.Println("tablename: " + tablename + " - " + err.Error())
	} else {
		if sqlIndex != "" {
			config.Get().Postgres.Exec(sqlIndex)
		}
	}
	return fmt.Sprintf("Table %s created", reflectType.Name())
}

func sqlCreateFields(reflectType reflect.Type, tablename, sqlCreate, sqlIndex string) (string, string) {
	indexFmt := "\ncreate %s " + tablename + "_%s on " + tablename + " (%s);"
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		tag := field.Tag.Get("sql")
		fieldName := strings.ToLower(field.Name)
		fieldType := sqlTypes[strings.ToLower(field.Type.Name())]

		if fieldType == "" {
			if field.Name == "Fields" {
				sqlCreate, sqlIndex = sqlCreateFields(field.Type, tablename, sqlCreate, sqlIndex)
			}
			continue
		}

		if fieldName != "id" {
			defaultValue := ""
			switch fieldType {
			case "bool":
				defaultValue = "DEFAULT false"
			case "timestamp":
				defaultValue = "DEFAULT current_timestamp"
			case "text":
				defaultValue = "DEFAULT ''"
			case "float", "float64":
				defaultValue = "DEFAULT 0.0"
			case "int", "int8":
				defaultValue = "DEFAULT 0"
			}
			sqlCreate += fmt.Sprintf("%s %s %s", fieldName, fieldType, defaultValue)
		}

		switch tag {
		case "pk":
			if fieldName == "id" {
				sqlCreate += "id SERIAL PRIMARY KEY"
			}
		case "index", "unique index":
			sqlIndex += fmt.Sprintf(indexFmt, tag, fieldName, fieldName)
		}
		sqlCreate += ", "
	}
	// return sqlCreate, sqlIndex
	return sqlCreate, ""
}

func SQLBulkInsert(table Tables, tableMapSlice []map[string]interface{}) string {
	reflectType := reflect.TypeOf(table).Elem()
	tablename := strings.ToLower(reflectType.Name())
	var sqlFields, sqlValues, sqlInsertBulk string

	for index, tableMap := range tableMapSlice {

		if tableMap["Createdate"] == nil {
			tableMap["Createdate"] = time.Now().Format(time.RFC3339)
		} else if fmt.Sprintf("%v", tableMap["Createdate"]) == "0001-01-01T00:00:00Z" {
			tableMap["Createdate"] = time.Now().Format(time.RFC3339)
		}

		if tableMap["Updatedate"] == nil {
			tableMap["Updatedate"] = time.Now().Format(time.RFC3339)
		} else if fmt.Sprintf("%v", tableMap["Updatedate"]) == "0001-01-01T00:00:00Z" {
			tableMap["Updatedate"] = time.Now().Format(time.RFC3339)
		}

		sqlFields, sqlValues = sqlBulkInsertFields(" (", " (", reflectType, tableMap)
		if sqlFields != "" && sqlValues != "" {
			if index%160 == 0 && index < len(tableMap)-1 {
				if sqlInsertBulk != "" {
					sqlInsertBulk = strings.TrimSuffix(sqlInsertBulk, "), ") + ");"
				}
				sqlInsertBulk += fmt.Sprintf("insert into %s %s) VALUES ",
					tablename, strings.TrimSuffix(sqlFields, ", "))
			}
			sqlInsertBulk += strings.TrimSuffix(sqlValues, ", ") + "), "
		}

	}
	return strings.TrimSuffix(sqlInsertBulk, "), ") + "); "
}

func sqlBulkInsertFields(sqlFields, sqlValues string, reflectType reflect.Type,
	tableMap map[string]interface{}) (string, string) {
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if tableMap[field.Name] != nil || field.Name == "Fields" {
			fieldName := strings.ToLower(field.Name)
			fieldType := sqlTypes[strings.ToLower(field.Type.Name())]
			if fieldType == "" || fieldName == "id" {
				if fieldName == "fields" {
					sqlFields, sqlValues = sqlBulkInsertFields(sqlFields, sqlValues, field.Type, tableMap)
				}
				continue
			}

			switch strings.ToLower(field.Type.Name()) {
			case "int", "int64", "uint", "uint64":
				tableMapFieldType := reflect.TypeOf(tableMap[field.Name])
				switch tableMapFieldType.Kind() {
				case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
					sqlValues += fmt.Sprintf("%d, ", tableMap[field.Name])
				default:
					sqlValues += fmt.Sprintf("%.f, ", tableMap[field.Name])
				}
			case "string", "time":
				sqlValues += fmt.Sprintf("'%v', ", strings.Replace(strings.Replace(tableMap[field.Name].(string), "'", "", -1), `"`, ``, -1))
			default: //bool and float
				sqlValues += fmt.Sprintf("%v, ", tableMap[field.Name])
			}
			sqlFields += fieldName + ", "
		}
	}
	return sqlFields, sqlValues
}

func (fi Fields) sqlInsert(table Tables, tableMap map[string]interface{}) (string, []interface{}) {
	if tableMap["ID"] != nil {
		delete(tableMap, "ID")
	}

	if tableMap["Workflow"] == nil || tableMap["Workflow"].(string) == "" {
		tableMap["Workflow"] = "draft"
	}

	if len(tableMap) == 0 || tableMap["Workflow"] == nil {
		return "", nil
	}

	if tableMap["Workflow"].(string) == "" {
		return "", nil
	}

	tableMap["Createdate"] = time.Now().Format(time.RFC3339)
	tableMap["Updatedate"] = time.Now().Format(time.RFC3339)

	reflectType := reflect.TypeOf(table).Elem()
	tablename := strings.ToLower(reflectType.Name())

	var sqlParams []interface{}
	sqlFields, sqlValues := " (", " ("
	sqlInsert := "insert into " + tablename
	sqlFields, sqlValues, sqlParams = sqlInsertFields(sqlFields, sqlValues, reflectType, tableMap, sqlParams)

	sqlInsert += strings.TrimSuffix(sqlFields, ", ") + " ) "
	sqlInsert += " VALUES "
	sqlInsert += strings.TrimSuffix(sqlValues, ", ") + " ) "
	sqlInsert += " RETURNING id"
	return sqlInsert, sqlParams
}

func sqlInsertFields(sqlFields, sqlValues string, reflectType reflect.Type,
	tableMap map[string]interface{}, sqlParams []interface{}) (string, string, []interface{}) {
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if tableMap[field.Name] != nil || field.Name == "Fields" {
			fieldName := strings.ToLower(field.Name)
			fieldType := sqlTypes[strings.ToLower(field.Type.Name())]
			if fieldType == "" {
				if fieldName == "fields" {
					sqlFields, sqlValues, sqlParams = sqlInsertFields(sqlFields, sqlValues, field.Type, tableMap, sqlParams)
				}
				continue
			}

			switch strings.ToLower(field.Type.Name()) {
			case "int", "int64", "uint", "uint64":
				tableMapFieldType := reflect.TypeOf(tableMap[field.Name])
				switch tableMapFieldType.Kind() {
				case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
					sqlParams = append(sqlParams, fmt.Sprintf("%d", tableMap[field.Name]))
				default:
					sqlParams = append(sqlParams, fmt.Sprintf("%.f", tableMap[field.Name]))
				}
			default:
				sqlParams = append(sqlParams, tableMap[field.Name])
			}

			sqlFields += fieldName + ", "
			sqlValues += fmt.Sprintf("$%v, ", len(sqlParams))
		}
	}
	return sqlFields, sqlValues, sqlParams
}

func (fi Fields) sqlUpdate(table Tables, tableMap map[string]interface{}) (string, []interface{}) {
	if tableMap["ID"] == nil {
		return "", nil
	}

	tableMap["Updatedate"] = time.Now().Format(time.RFC3339)

	reflectType := reflect.TypeOf(table).Elem()
	tablename := strings.ToLower(reflectType.Name())

	var sqlParams []interface{}
	sqlUpdate := "update " + tablename + " set "
	sqlUpdate, sqlParams = sqlUpdateFields(sqlUpdate, reflectType, tableMap, sqlParams)
	sqlUpdate = strings.TrimSuffix(sqlUpdate, ", ")

	sqlIDfmt := "%.f"
	switch reflect.TypeOf(tableMap["ID"]).Kind() {
	case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
		sqlIDfmt = "%d"
	}

	sqlParams = append(sqlParams, fmt.Sprintf(sqlIDfmt, tableMap["ID"]))
	sqlUpdate += fmt.Sprintf(" where id = $%v", len(sqlParams))

	return sqlUpdate, sqlParams
}

func sqlUpdateFields(sqlUpdate string, reflectType reflect.Type, tableMap map[string]interface{}, sqlParams []interface{}) (string, []interface{}) {
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)

		if tableMap[field.Name] != nil || field.Name == "Fields" {
			fieldName := strings.ToLower(field.Name)
			fieldType := sqlTypes[strings.ToLower(field.Type.Name())]

			if fieldType == "" || fieldName == "id" {
				if fieldName == "fields" {
					sqlUpdate, sqlParams = sqlUpdateFields(sqlUpdate, field.Type, tableMap, sqlParams)
				}
				continue
			}

			switch strings.ToLower(field.Type.Name()) {
			case "int", "int64", "uint", "uint64":
				tableMapFieldType := reflect.TypeOf(tableMap[field.Name])
				switch tableMapFieldType.Kind() {
				case reflect.Int, reflect.Int64, reflect.Uint, reflect.Uint64:
					sqlParams = append(sqlParams, fmt.Sprintf("%d", tableMap[field.Name]))
				default:
					sqlParams = append(sqlParams, fmt.Sprintf("%.f", tableMap[field.Name]))
				}
			default:
				sqlParams = append(sqlParams, tableMap[field.Name])
			}

			sqlUpdate += fmt.Sprintf("%s = $%v, ", fieldName, len(sqlParams))
		}
	}
	return sqlUpdate, sqlParams
}

func (fi Fields) sqlSelect(table Tables, tableMap map[string]interface{}, searchParams *SearchParams) (string, []interface{}) {

	reflectType := reflect.TypeOf(table).Elem()
	tablename := strings.ToLower(reflectType.Name())

	var fields []string

	for fieldName := range tableMap {
		columnName := fmt.Sprintf("%s.%s as %s", tablename, fieldName, fieldName)
		fields = append(fields, columnName)
	}
	sqlSelect := fmt.Sprintf("select %s from %s where ", strings.Join(fields, ", "), tablename)

	var sqlParams []interface{}
	if searchParams.Workflow != "" {
		sqlParams = append(sqlParams, searchParams.Workflow)
		sqlSelect += fmt.Sprintf("Workflow = $%v and ", len(sqlParams))
	}

	//Take Filter fields and loop through
	if searchParams.Filter != nil && len(searchParams.Filter) > 0 {
		for fieldName, fieldValue := range searchParams.Filter {
			fieldValue = strings.TrimSpace(fieldValue)
			if len(fieldValue) > 0 {
				fieldValue = "%" + strings.ToLower(fieldValue) + "%"
				sqlParams = append(sqlParams, fieldValue)
				if !strings.Contains(fieldName, ".") {
					fieldName = fmt.Sprintf("%s.%s", tablename, fieldName)
				}
				if strings.Contains(strings.ToLower(fieldName), "date") {
					sqlSelect += fmt.Sprintf("%s like $%v and ", strings.ToLower(fieldName), len(sqlParams))
				} else {
					sqlSelect += fmt.Sprintf("lower(%s) like lower($%v) and ", strings.ToLower(fieldName), len(sqlParams))
				}
			}
		}
	}
	//loop through extra fields and values

	return sqlSelect, sqlParams
}
