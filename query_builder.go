package cube

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type columnStructMapping map[string]string

type QueryBuilder struct {
	table            string
	selects          []string
	wheres           []string
	args             []interface{}
	inserts          []string
	orderByArgs      []string
	orderByDirection string
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

func (qb *QueryBuilder) Select(selects ...string) *QueryBuilder {
	if len(selects) == 0 {
		return qb
	}
	for _, v := range selects {
		qb.selects = append(qb.selects, parenthesesWrap(v))
	}
	return qb
}

func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.table = table
	return qb
}

// Where clause in the form of `...Where("x >= ?", y)`
func (qb *QueryBuilder) Where(query string, args ...interface{}) *QueryBuilder {
	qb.wheres = append(qb.wheres, query)
	qb.args = append(qb.args, args...)
	return qb
}

func (qb *QueryBuilder) OrderBy(desc bool, cols ...string) *QueryBuilder {
	qb.orderByArgs = append(qb.orderByArgs, cols...)
	qb.orderByDirection = " ASC "
	if desc {
		qb.orderByDirection = " DESC "
	}
	return qb
}

// Select columns from a table based on the interface that is passed into it
func (qb *QueryBuilder) SelectStruct(obj interface{}) *QueryBuilder {
	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		qb.Select(field.Name)
	}

	return qb
}

func (qb *QueryBuilder) GetStatement() string {
	var query string
	if len(qb.selects) > 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(qb.selects, ","), parenthesesWrap(qb.table))
		if len(qb.wheres) > 0 {
			query += " WHERE " + strings.Join(qb.wheres, " AND ")
		}

		if len(qb.orderByArgs) > 0 {
			query += " ORDER BY " + strings.Join(qb.orderByArgs, ", ") + qb.orderByDirection
		}
		return query
	}

	if len(qb.inserts) > 0 {
		values := make([]string, len(qb.inserts))
		for i := range values {
			values[i] = "?"
		}

		var vBlock []string
		numRows := len(qb.args) / len(qb.inserts)
		for i := 0; i < numRows; i++ {
			vBlock = append(vBlock, fmt.Sprintf("(%s)", strings.Join(values, ",")))
		}

		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", parenthesesWrap(qb.table), strings.Join(qb.inserts, ","), strings.Join(vBlock, ","))
	}
	
	return query
}

func (qb *QueryBuilder) Exec(db *sql.DB) (sql.Result, error) {
	return db.Exec(qb.GetStatement(), qb.args...)
}

func (qb *QueryBuilder) Query(db *sql.DB) (*sql.Rows, error) {
	query := qb.GetStatement()
	return db.Query(query, qb.args...)
}

// Scan the sql.Rows into the passed struct
// Some known limitations:
// Column names need to match the struct properties exactly or it will zero the value
func ScanStruct(rows *sql.Rows, out interface{}) error {
	colMap := make(columnStructMapping)
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	for i := range values {
		var field interface{}
		values[i] = &field
	}

	var reflectValue reflect.Value

	// Check if out is a slice
	outValue := reflect.ValueOf(out)
	if outValue.Kind() == reflect.Ptr && outValue.Elem().Kind() == reflect.Slice {
		reflectValue = outValue.Elem()
	} else {
		return errors.New("out must be a pointer to a slice")
	}

	fieldsPtr := reflect.New(reflectValue.Type().Elem())
	for i := 0; i < fieldsPtr.Elem().NumField(); i++ {
		fieldName := fieldsPtr.Elem().Type().Field(i).Name
		for _, v := range columns {
			if strings.EqualFold(v, fieldName) {
				colMap[v] = fieldName
			}
		}
	}

	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return err
		}

		structPtr := reflect.New(reflectValue.Type().Elem())
		structVal := structPtr.Elem()

		for i, column := range columns {
			field := structVal.FieldByName(colMap[column])
			if !field.IsValid() {
				continue
			}

			value := reflect.ValueOf(values[i]).Elem().Interface()
			fieldType := field.Type()

			switch fieldType.Kind() {
			case reflect.Slice:
				sliceType := fieldType.Elem()
				slice := reflect.MakeSlice(fieldType, 1, 1)

				if sliceType.Kind() == reflect.Uint8 {
					slice = reflect.ValueOf(value)
				}

				field.Set(slice)
			case reflect.String:
				if value != nil {
					field.Set(reflect.ValueOf(value).Convert(fieldType))
				}
			default:
				if value != nil {
					field.Set(reflect.ValueOf(value).Convert(fieldType))
				}
			}
		}
		reflectValue = reflect.Append(reflectValue, structVal)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	outValue.Elem().Set(reflectValue)

	return nil
}

func parenthesesWrap(str string) string {
	return fmt.Sprintf("[%s]", str)
}
