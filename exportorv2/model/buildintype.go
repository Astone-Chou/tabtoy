package model

import (
	"github.com/davyxu/tabtoy/proto/tool"
)

type BuildInTypeKind int

const (
	BuildInTypeKind_None BuildInTypeKind = iota
	BuildInTypeKind_Enum
	BuildInTypeKind_Struct
)

type BuildInType struct {
	Name string
	Kind BuildInTypeKind

	FieldByName   map[string]*FieldDefine
	FieldByNumber map[int32]*FieldDefine
	Fields        []*FieldDefine
}

func (self *BuildInType) Add(def *FieldDefine) {

	self.FieldByName[def.Name] = def
	self.FieldByNumber[def.EnumValue] = def
	self.Fields = append(self.Fields, def)
}

func (self *BuildInType) FieldByValueAndMeta(value string) *FieldDefine {

	for _, v := range self.FieldByName {

		if v.Name == value {
			return v
		}

		if v.Meta.Alias == value {
			return v
		}

	}

	return nil
}

func NewBuildInType() *BuildInType {
	return &BuildInType{
		FieldByName:   make(map[string]*FieldDefine),
		FieldByNumber: make(map[int32]*FieldDefine),
	}
}

type BuildInTypeSet struct {
	TypeByName map[string]*BuildInType
	Types      []*BuildInType

	Pragma tool.BuildInTypePragmaV2
}

func (self *BuildInTypeSet) Add(def *BuildInType) {

	self.Types = append(self.Types, def)
	self.TypeByName[def.Name] = def
}

func NewBuildInTypeSet() *BuildInTypeSet {
	return &BuildInTypeSet{
		TypeByName: make(map[string]*BuildInType),
	}
}
