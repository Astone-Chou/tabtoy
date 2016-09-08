package exportorv2

import (
	"strconv"

	"github.com/davyxu/tabtoy/exportorv2/model"
	"github.com/davyxu/tabtoy/util"
	"github.com/golang/protobuf/proto"
)

const (
	// 信息所在的行
	TypeSheetRow_Pragma    = 0 // 配置
	TypeSheetRow_Comment   = 1 // 字段名(对应proto)
	TypeSheetRow_DataBegin = 2 // 数据开始
)

const (
	// 信息所在列
	TypeSheetCol_ObjectType = 0 // 对象类型
	TypeSheetCol_FieldName  = 1 // 字段名
	TypeSheetCol_FieldType  = 2 // 字段名
	TypeSheetCol_Value      = 3 // 值
	TypeSheetCol_Meta       = 4 // 特性
)

type TypeSheet struct {
	*Sheet

	*model.BuildInTypeSet
}

func (self *TypeSheet) Parse() bool {

	// 是否继续读行
	var readingLine bool = true

	var td *model.BuildInType

	rawPragma := self.GetCellData(TypeSheetRow_Pragma, 0)

	if err := proto.UnmarshalText(rawPragma, &self.BuildInTypeSet.Pragma); err != nil {
		self.Row = TypeSheetRow_Pragma
		self.Column = 0
		log.Errorf("parse pragma failed: %s", rawPragma)
		goto ErrorStop
	}

	// 遍历每一行
	for self.Row = TypeSheetRow_DataBegin; readingLine; self.Row++ {

		// ====================解析对象类型====================
		// 第一列是空的，结束
		if self.GetCellData(self.Row, TypeSheetCol_ObjectType) == "" {
			break
		}

		var fd model.FieldDefine

		rawTypeName := self.GetCellData(self.Row, TypeSheetCol_ObjectType)

		existType, ok := self.BuildInTypeSet.TypeByName[rawTypeName]

		if ok {

			td = existType

		} else {

			td = model.NewBuildInType()
			td.Name = rawTypeName
			self.BuildInTypeSet.Add(td)
		}

		// ====================解析字段名====================
		fd.Name = self.GetCellData(self.Row, TypeSheetCol_FieldName)

		// ====================解析字段名====================
		rawFieldType := self.GetCellData(self.Row, TypeSheetCol_FieldType)

		// 解析普通类型
		if ft, ok := model.ParseFieldType(rawFieldType); ok {
			fd.Type = ft
		} else {

			// 解析内建类型
			if buildinType, ok := self.BuildInTypeSet.TypeByName[rawFieldType]; ok {

				// 只有枚举( 结构体不允许再次嵌套, 增加理解复杂度 )
				if buildinType.Kind != model.BuildInTypeKind_Enum {
					self.Column = TypeSheetCol_FieldType
					log.Errorln("struct field can only be normal type and enum", rawFieldType)
					goto ErrorStop
				}

				fd.Type = model.FieldType_Enum
				fd.BuildInType = buildinType

			} else {

				self.Column = TypeSheetCol_FieldType
				log.Errorln("unknown field type: ", rawFieldType)
				goto ErrorStop
			}

		}

		// ====================解析值====================
		rawValue := self.GetCellData(self.Row, TypeSheetCol_Value)

		var kind model.BuildInTypeKind

		// 非空值是枚举
		if rawValue != "" {

			// 解析枚举值
			if v, err := strconv.Atoi(rawValue); err == nil {
				fd.EnumValue = int32(v)
			} else {
				self.Column = TypeSheetCol_Value
				log.Errorln("parse type value failed:", err)
				goto ErrorStop
			}
			kind = model.BuildInTypeKind_Enum
		} else {
			kind = model.BuildInTypeKind_Struct
		}

		if td.Kind == model.BuildInTypeKind_None {
			td.Kind = kind
			// 一些字段有填值, 一些没填值
		} else if td.Kind != kind {
			self.Column = TypeSheetCol_Value
			log.Errorln("buildin kind shold be same", td.Kind, kind)
			goto ErrorStop
		}

		// ====================解析特性====================
		metaString := self.GetCellData(self.Row, TypeSheetCol_Meta)

		if err := proto.UnmarshalText(metaString, &fd.Meta); err != nil {
			log.Errorln("parse field header failed", err)
			return false
		}

		td.Add(&fd)

	}

	return self.checkProtobufCompatibility()

ErrorStop:

	r, c := self.GetRC()

	log.Errorf("%s|%s(%s)", self.file.FileName, self.Name, util.ConvR1C1toA1(r, c))
	return false
}

// 检查protobuf兼容性
func (self *TypeSheet) checkProtobufCompatibility() bool {

	for _, bt := range self.BuildInTypeSet.Types {
		if bt.Kind == model.BuildInTypeKind_Enum {

			// proto3 需要枚举有0值
			if _, ok := bt.FieldByNumber[0]; !ok {
				log.Errorf("proto3 require enum has value 0 in '%s'", bt.Name)
				return false
			}
		}
	}

	return true
}

func newTypeSheet(sheet *Sheet) *TypeSheet {
	return &TypeSheet{
		Sheet:          sheet,
		BuildInTypeSet: model.NewBuildInTypeSet(),
	}
}
