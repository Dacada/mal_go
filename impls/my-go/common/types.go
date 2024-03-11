package common;

type MalType interface {}

type MalTypeList []MalType
type MalTypeVector []MalType
type MalTypeHashMap map[MalType]MalType
type MalTypeInteger int64
type MalTypeNil int
type MalTypeBoolean bool
type MalTypeKeyword string
type MalTypeString string
type MalTypeSymbol string
type MalTypeFunction func([]MalType) (MalType, error)
