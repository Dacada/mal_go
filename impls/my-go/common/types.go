package common;

type MalType interface {
	IsMalType() bool  // boogus function so this interface only includes these types
}

type MalTypeList struct {
	List []MalType
	Meta MalType
}
type MalTypeVector struct {
	Vector []MalType
	Meta MalType
}
type MalTypeHashMap struct {
	HashMap map[MalType]MalType
	Meta MalType
}
type MalTypeInteger int64
type MalTypeNil struct {}
type MalTypeBoolean bool
type MalTypeKeyword string
type MalTypeString string
type MalTypeSymbol string
type MalTypeFunction struct {
	Func func([]MalType) (MalType, error)
	Meta MalType
}
type MalTypeTCOFunction struct {
	Ast MalType
	Params []MalTypeSymbol
	Env Env
	IsMacro bool
	Fn MalTypeFunction
}
type MalTypeAtom struct {
	Ptr *MalType
}
type MalTypeError struct {
	InnerMalType MalType
}
func (x MalTypeError) Error() string {
	return PrStr(x.InnerMalType, false)
}

func NewMalList(list []MalType) MalTypeList {
	return MalTypeList{list, MalTypeNil{}}
}
func NewMalVector(vector []MalType) MalTypeVector {
	return MalTypeVector{vector, MalTypeNil{}}
}
func NewMalHashMap(hashmap map[MalType]MalType) MalTypeHashMap {
	return MalTypeHashMap{hashmap, MalTypeNil{}}
}
func NewMalFunction(function func([]MalType) (MalType, error)) MalTypeFunction {
	return MalTypeFunction{function, MalTypeNil{}}
}

// implementation of the bogus interface for each mal type so that they can be considered members of the type
func (_ MalTypeList) IsMalType() bool {
	return true
}
func (_ MalTypeVector) IsMalType() bool {
	return true
}
func (_ MalTypeHashMap) IsMalType() bool {
	return true
}
func (_ MalTypeInteger) IsMalType() bool {
	return true
}
func (_ MalTypeNil) IsMalType() bool {
	return true
}
func (_ MalTypeBoolean) IsMalType() bool {
	return true
}
func (_ MalTypeKeyword) IsMalType() bool {
	return true
}
func (_ MalTypeString) IsMalType() bool {
	return true
}
func (_ MalTypeSymbol) IsMalType() bool {
	return true
}
func (_ MalTypeFunction) IsMalType() bool {
	return true
}
func (_ MalTypeTCOFunction) IsMalType() bool {
	return true
}
func (_ MalTypeAtom) IsMalType() bool {
	return true
}
func (_ MalTypeError) IsMalType() bool {
	return true
}
