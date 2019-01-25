package database

var sqlTypes = map[string]string{
	"bool":    "bool",
	"time":    "timestamp",
	"string":  "text",
	"int":     "int",
	"uint":    "int",
	"int64":   "int8",
	"uint32":  "int8",
	"uint64":  "int8",
	"float32": "float8",
	"float64": "float8",
}

type Tables interface {
	ToMap() (mapInterface map[string]interface{})
	FillStruct(tableMap map[string]interface{}) error
}

var AllTables = map[string]Tables{
	"Candidates": &Candidates{},
	"Documents":  &Documents{},
	"Positions":  &Positions{},
	"Profiles":   &Profiles{},
	"Proposals":  &Proposals{},
	"Settings":   &Settings{},
	"Users":      &Users{},
	"Voters":     &Voters{},
	"Votes":      &Votes{},
}

func createTable(tableName string) (Message []string) {
	switch tableName {
	default:
		tableName = tableName[1:]
		if AllTables[tableName] != nil {
			Message = append(Message, Fields{}.sqlCreate(AllTables[tableName]))
			switch tableName {
			case "Users":
				users := new(Users)
				users.Setup()
			}
		} else {
			Message = append(Message, "Please Specify Table")
		}

	case "/all":
		for tableName, table := range AllTables {
			Message = append(Message, Fields{}.sqlCreate(table))
			switch tableName {
			case "Users":
				users := new(Users)
				users.Setup()
			}
		}
	}
	return Message
}

// SearchParams serves as default parameters used in generating sql prepared statements
type SearchParams struct {
	Field, Text, Workflow string

	Skip, Limit uint64

	UserID,
	ID, RefID uint64
	RefField string

	OrderBy  string
	OrderAsc bool

	Filter map[string]string
}
