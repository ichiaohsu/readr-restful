package mysql

type SQLsv struct {
	Statement string
	Variable  interface{}
}

type Sqlfield struct {
	Table   string
	Pattern string
	Fields  []string
}

// SQLO , not SOLO, stands for "SQL Object".
// It tries to mapping SQL statement to struct
type SQLO struct {

	// Table hosts the table name for select
	Table string

	// Fields hosts all the fields that will appear in SELECT statement
	Fields []Sqlfield

	// Join provide table to be joined, in string form
	Join []string

	// Where comprises SQL statements strings in 'WHERE' section
	Where []SQLsv

	// Order by maps to the ORDER BY section in SQL
	Orderby string

	// Pagination is the limit string, in this pattern: LIMIT [max_results] OFFSET [page-1]
	Pagination string

	// Args comprises all the argument corresponding to placeholders in SQL statements
	// They will be passed into sqlx functions
	Args []interface{}
}

// NewSQLO generates a new SQLO object pointer with given options
func NewSQLO(options ...func(*SQLO)) *SQLO {
	so := SQLO{}
	for _, option := range options {
		option(&so)
	}
	return &so
}
