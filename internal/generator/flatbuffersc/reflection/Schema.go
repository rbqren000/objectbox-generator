// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package reflection

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Schema struct {
	_tab flatbuffers.Table
}

const SchemaIdentifier = "BFBS"

func GetRootAsSchema(buf []byte, offset flatbuffers.UOffsetT) *Schema {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Schema{}
	x.Init(buf, n+offset)
	return x
}

func FinishSchemaBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	identifierBytes := []byte(SchemaIdentifier)
	builder.FinishWithFileIdentifier(offset, identifierBytes)
}

func SchemaBufferHasIdentifier(buf []byte) bool {
	return flatbuffers.BufferHasIdentifier(buf, SchemaIdentifier)
}

func GetSizePrefixedRootAsSchema(buf []byte, offset flatbuffers.UOffsetT) *Schema {
	n := flatbuffers.GetUOffsetT(buf[offset+flatbuffers.SizeUint32:])
	x := &Schema{}
	x.Init(buf, n+offset+flatbuffers.SizeUint32)
	return x
}

func FinishSizePrefixedSchemaBuffer(builder *flatbuffers.Builder, offset flatbuffers.UOffsetT) {
	identifierBytes := []byte(SchemaIdentifier)
	builder.FinishSizePrefixedWithFileIdentifier(offset, identifierBytes)
}

func SizePrefixedSchemaBufferHasIdentifier(buf []byte) bool {
	return flatbuffers.SizePrefixedBufferHasIdentifier(buf, SchemaIdentifier)
}

func (rcv *Schema) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Schema) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Schema) Objects(obj *Object, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Schema) ObjectsByKey(obj *Object, key string) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Vector(o)
		return obj.LookupByKey(key, x, rcv._tab.Bytes)
	}
	return false
}

func (rcv *Schema) ObjectsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Schema) Enums(obj *Enum, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Schema) EnumsByKey(obj *Enum, key string) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Vector(o)
		return obj.LookupByKey(key, x, rcv._tab.Bytes)
	}
	return false
}

func (rcv *Schema) EnumsLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Schema) FileIdent() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Schema) FileExt() []byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	}
	return nil
}

func (rcv *Schema) RootTable(obj *Object) *Object {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(Object)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *Schema) Services(obj *Service, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Schema) ServicesByKey(obj *Service, key string) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		x := rcv._tab.Vector(o)
		return obj.LookupByKey(key, x, rcv._tab.Bytes)
	}
	return false
}

func (rcv *Schema) ServicesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

func (rcv *Schema) AdvancedFeatures() AdvancedFeatures {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		return AdvancedFeatures(rcv._tab.GetUint64(o + rcv._tab.Pos))
	}
	return 0
}

func (rcv *Schema) MutateAdvancedFeatures(n AdvancedFeatures) bool {
	return rcv._tab.MutateUint64Slot(16, uint64(n))
}

/// All the files used in this compilation. Files are relative to where
/// flatc was invoked.
func (rcv *Schema) FbsFiles(obj *SchemaFile, j int) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	}
	return false
}

func (rcv *Schema) FbsFilesByKey(obj *SchemaFile, key string) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		x := rcv._tab.Vector(o)
		return obj.LookupByKey(key, x, rcv._tab.Bytes)
	}
	return false
}

func (rcv *Schema) FbsFilesLength() int {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(18))
	if o != 0 {
		return rcv._tab.VectorLen(o)
	}
	return 0
}

/// All the files used in this compilation. Files are relative to where
/// flatc was invoked.
func SchemaStart(builder *flatbuffers.Builder) {
	builder.StartObject(8)
}
func SchemaAddObjects(builder *flatbuffers.Builder, objects flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(objects), 0)
}
func SchemaStartObjectsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func SchemaAddEnums(builder *flatbuffers.Builder, enums flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(enums), 0)
}
func SchemaStartEnumsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func SchemaAddFileIdent(builder *flatbuffers.Builder, fileIdent flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(fileIdent), 0)
}
func SchemaAddFileExt(builder *flatbuffers.Builder, fileExt flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(fileExt), 0)
}
func SchemaAddRootTable(builder *flatbuffers.Builder, rootTable flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(rootTable), 0)
}
func SchemaAddServices(builder *flatbuffers.Builder, services flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(services), 0)
}
func SchemaStartServicesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func SchemaAddAdvancedFeatures(builder *flatbuffers.Builder, advancedFeatures AdvancedFeatures) {
	builder.PrependUint64Slot(6, uint64(advancedFeatures), 0)
}
func SchemaAddFbsFiles(builder *flatbuffers.Builder, fbsFiles flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(7, flatbuffers.UOffsetT(fbsFiles), 0)
}
func SchemaStartFbsFilesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	return builder.StartVector(4, numElems, 4)
}
func SchemaEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
