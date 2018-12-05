/*
 * Copyright 2018 ObjectBox Ltd. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package templates

import (
	"text/template"
)

var BindingTemplate = template.Must(template.New("binding").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

package {{.Binding.Package}}

import (
	"github.com/google/flatbuffers/go"
	"github.com/objectbox/objectbox-go/objectbox"
	"github.com/objectbox/objectbox-go/objectbox/fbutils"
	{{if .Binding.UsesStrconv}}"strconv"{{end}}
)

{{range $entity := .Binding.Entities -}}
{{$entityNameCamel := $entity.Name | StringCamel -}}
type {{$entityNameCamel}}_EntityInfo struct {
	Id objectbox.TypeId
	Uid uint64
}

var {{$entity.Name}}Binding = {{$entityNameCamel}}_EntityInfo {
	Id: {{$entity.Id}}, 
	Uid: {{$entity.Uid}},
}

// {{$entity.Name}}_ contains type-based Property helpers to facilitate some common operations such as Queries. 
var {{$entity.Name}}_ = struct {
	{{range $property := $entity.Properties -}}
    {{$property.Name}} *objectbox.
		{{- if and (eq $entity.IdProperty.Name $property.Name) (eq $entity.IdProperty.GoType "string")}}PropertyStringUint64
		{{- else}}Property{{$property.GoType | TypeIdentifier}}
		{{- end}}
    {{end -}}
}{
	{{range $property := $entity.Properties -}}
    {{$property.Name}}: &objectbox.
			{{- if and (eq $entity.IdProperty.Name $property.Name) (eq $entity.IdProperty.GoType "string")}}PropertyStringUint64
			{{- else}}Property{{$property.GoType | TypeIdentifier}}
			{{- end}}{
		Property: &objectbox.Property{
			Id: {{$property.Id}},
		},
	},
    {{end -}}
}

// GeneratorVersion is called by the ObjectBox to verify the compatibility of the generator used to generate this code	
func ({{$entityNameCamel}}_EntityInfo) GeneratorVersion() int {
	return {{$.GeneratorVersion}}
}

// AddToModel is called by the ObjectBox during model build
func ({{$entityNameCamel}}_EntityInfo) AddToModel(model *objectbox.Model) {
    model.Entity("{{$entity.Name}}", {{$entity.Id}}, {{$entity.Uid}})
    {{range $property := $entity.Properties -}}
    model.Property("{{$property.ObName}}", objectbox.PropertyType_{{$property.ObType}}, {{$property.Id}}, {{$property.Uid}})
    {{if len $property.ObFlags -}}
        model.PropertyFlags(
        {{- range $key, $flag := $property.ObFlags -}}
            {{if gt $key 0}} | {{end}}objectbox.PropertyFlags_{{$flag}}
        {{- end}})
        {{- /* TODO model.PropertyIndexId() && model.PropertyRelation() */}}
    {{end -}}
	{{if $property.Relation}}model.PropertyRelation("{{$property.Relation.Target}}", {{$property.Index.Id}}, {{$property.Index.Uid}})
	{{else if $property.Index}}model.PropertyIndex({{$property.Index.Id}}, {{$property.Index.Uid}})
    {{end -}}
    {{end -}}
    model.EntityLastPropertyId({{$entity.LastPropertyId.GetId}}, {{$entity.LastPropertyId.GetUid}})
}

// GetId is called by the ObjectBox during Put operations to check for existing ID on an object
func ({{$entityNameCamel}}_EntityInfo) GetId(object interface{}) (uint64, error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	if len(object.(*{{$entity.Name}}).{{$entity.IdProperty.Name}}) == 0 {
		return 0, nil
	} else {
		return strconv.ParseUint(object.(*{{$entity.Name}}).{{$entity.IdProperty.Name}}, 10, 64)
	}
	{{- else -}}
	return object.(*{{$entity.Name}}).{{$entity.IdProperty.Name}}, nil
	{{- end}}
}

// SetId is called by the ObjectBox during Put to update an ID on an object that has just been inserted
func ({{$entityNameCamel}}_EntityInfo) SetId(object interface{}, id uint64) error {
	{{if eq $entity.IdProperty.GoType "string" -}}
	object.(*{{$entity.Name}}).Id = strconv.FormatUint(id, 10)
	{{- else -}}
	object.(*{{$entity.Name}}).{{$entity.IdProperty.Name}} = id
	{{- end}}
	return nil
}

// Flatten is called by the ObjectBox to transform an object to a FlatBuffer
func ({{$entityNameCamel}}_EntityInfo) Flatten(object interface{}, fbb *flatbuffers.Builder, id uint64) {
    {{if $entity.HasNonIdProperty}}obj := object.(*{{$entity.Name}}){{end -}}

    {{- range $property := $entity.Properties}}
        {{- if eq $property.FbType "UOffsetT"}}
            {{- if eq $property.GoType "string"}}
    var offset{{$property.Name}} = fbutils.CreateStringOffset(fbb, obj.{{$property.Name}})
            {{- else if eq $property.GoType "[]byte"}}
    var offset{{$property.Name}} = fbutils.CreateByteVectorOffset(fbb, obj.{{$property.Name}})
            {{- end -}}
        {{end}}
    {{- end}}

    // build the FlatBuffers object
    fbb.StartObject({{$entity.LastPropertyId.GetId}})
    {{range $property := $entity.Properties -}}
    fbutils.Set{{$property.FbType}}Slot(fbb, {{$property.FbSlot}},
        {{- if eq $property.FbType "UOffsetT"}} offset{{$property.Name}})
        {{- else if eq $property.Name $entity.IdProperty.Name}} id)
        {{- else if eq $property.GoType "int"}} int64(obj.{{$property.Name}}))
        {{- else if eq $property.GoType "uint"}} uint64(obj.{{$property.Name}}))
        {{- else}} obj.{{$property.Name}})
        {{- end}}
    {{end -}}
}

// ToObject is called by the ObjectBox to load an object from a FlatBuffer 
func ({{$entityNameCamel}}_EntityInfo) ToObject(bytes []byte) interface{} {
	table := &flatbuffers.Table{
		Bytes: bytes,
		Pos:   flatbuffers.GetUOffsetT(bytes),
	}

	return &{{$entity.Name}}{
	{{- range $property := $entity.Properties}}
		{{$property.Name}}: {{if eq $property.GoType "bool"}} table.GetBoolSlot({{$property.FbvTableOffset}}, false)
        {{- else if eq $property.GoType "int"}} int(table.GetUint64Slot({{$property.FbvTableOffset}}, 0))
        {{- else if eq $property.GoType "uint"}} uint(table.GetUint64Slot({{$property.FbvTableOffset}}, 0))
		{{- else if eq $property.GoType "rune"}} rune(table.GetInt32Slot({{$property.FbvTableOffset}}, 0))
		{{- else if and (eq $property.GoType "string") (eq $property.FbType "Uint64")}} strconv.FormatUint(table.GetUint64Slot({{$property.FbvTableOffset}}, 0), 10)
		{{- else if eq $property.GoType "string"}} fbutils.GetStringSlot(table, {{$property.FbvTableOffset}})
        {{- else if eq $property.GoType "[]byte"}} fbutils.GetByteVectorSlot(table, {{$property.FbvTableOffset}})
		{{- else}} table.Get{{$property.GoType | StringTitle}}Slot({{$property.FbvTableOffset}}, 0)
        {{- end}},
	{{- end}}
	}
}

// MakeSlice is called by the ObjectBox to construct a new slice to hold the read objects  
func ({{$entityNameCamel}}_EntityInfo) MakeSlice(capacity int) interface{} {
	return make([]*{{$entity.Name}}, 0, capacity)
}

// AppendToSlice is called by the ObjectBox to fill the slice of the read objects
func ({{$entityNameCamel}}_EntityInfo) AppendToSlice(slice interface{}, object interface{}) interface{} {
	return append(slice.([]*{{$entity.Name}}), object.(*{{$entity.Name}}))
}

// Box provides CRUD access to {{$entity.Name}} objects
type {{$entity.Name}}Box struct {
	*objectbox.Box
}

// BoxFor{{$entity.Name}} opens a box of {{$entity.Name}} objects 
func BoxFor{{$entity.Name}}(ob *objectbox.ObjectBox) *{{$entity.Name}}Box {
	return &{{$entity.Name}}Box{
		Box: ob.InternalBox({{$entity.Id}}),
	}
}

// Put synchronously inserts/updates a single object.
// In case the {{$entity.IdProperty.Name}} is not specified, it would be assigned automatically (auto-increment).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Name}} property on the passed object will be assigned the new ID as well.
func (box *{{$entity.Name}}Box) Put(object *{{$entity.Name}}) ({{$entity.IdProperty.GoType}}, error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	if id, err := box.Box.Put(object); err != nil {
		return "", err
	} else {
		return strconv.FormatUint(id, 10), nil
	}
	{{- else -}}
	return box.Box.Put(object)
	{{- end}}
}

// PutAsync asynchronously inserts/updates a single object.
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Name}} property on the passed object will be assigned the new ID as well.
// 
// It's executed on a separate internal thread for better performance.
//
// There are two main use cases:
//
// 1) "Put & Forget:" you gain faster puts as you don't have to wait for the transaction to finish.
//
// 2) Many small transactions: if your write load is typically a lot of individual puts that happen in parallel,
// this will merge small transactions into bigger ones. This results in a significant gain in overall throughput.
//
//
// In situations with (extremely) high async load, this method may be throttled (~1ms) or delayed (<1s).
// In the unlikely event that the object could not be enqueued after delaying, an error will be returned.
//
// Note that this method does not give you hard durability guarantees like the synchronous Put provides.
// There is a small time window (typically 3 ms) in which the data may not have been committed durably yet.
func (box *{{$entity.Name}}Box) PutAsync(object *{{$entity.Name}}) ({{$entity.IdProperty.GoType}}, error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	if id, err := box.Box.PutAsync(object); err != nil {
		return "", err
	} else {
		return strconv.FormatUint(id, 10), nil
	}
	{{- else -}}
	return box.Box.PutAsync(object)
	{{- end}}
}

// PutAll inserts multiple objects in single transaction.
// In case {{$entity.IdProperty.Name}}s are not set on the objects, they would be assigned automatically (auto-increment).
// 
// Returns: IDs of the put objects (in the same order).
// When inserting, the {{$entity.Name}}.{{$entity.IdProperty.Name}} property on the objects in the slice will be assigned the new IDs as well.
//
// Note: In case an error occurs during the transaction, some of the objects may already have the {{$entity.Name}}.{{$entity.IdProperty.Name}} assigned    
// even though the transaction has been rolled back and the objects are not stored under those IDs.
//
// Note: The slice may be empty or even nil; in both cases, an empty IDs slice and no error is returned.
func (box *{{$entity.Name}}Box) PutAll(objects []*{{$entity.Name}}) ([]{{$entity.IdProperty.GoType}}, error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	ids, err := box.Box.PutAll(objects)
	if err != nil || len(ids) == 0 {
		return []string{}, err
	}

	var stringIds = make([]string, len(ids))
	for i, id := range ids {
		stringIds[i] = strconv.FormatUint(id, 10)
	}

	return stringIds, nil 
	{{- else -}}
	return box.Box.PutAll(objects)
	{{- end}}
}

// Get reads a single object.
//
// Returns nil (and no error) in case the object with the given ID doesn't exist.
func (box *{{$entity.Name}}Box) Get(id {{$entity.IdProperty.GoType}}) (*{{$entity.Name}}, error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	idUint64, parseErr := strconv.ParseUint(id, 10, 64)
	if parseErr != nil {
		return nil, parseErr
	}

	object, err := box.Box.Get(idUint64)

	{{else -}}
	object, err := box.Box.Get(id)
	{{end -}}

	if err != nil {
		return nil, err
	} else if object == nil {
		return nil, nil
	}
	return object.(*{{$entity.Name}}), nil
}

// Get reads all stored objects
func (box *{{$entity.Name}}Box) GetAll() ([]*{{$entity.Name}}, error) {
	objects, err := box.Box.GetAll()
	if err != nil {
		return nil, err
	}
	return objects.([]*{{$entity.Name}}), nil
}

// Remove deletes a single object
func (box *{{$entity.Name}}Box) Remove(object *{{$entity.Name}}) (err error) {
	{{if eq $entity.IdProperty.GoType "string" -}}
	idUint64, parseErr := strconv.ParseUint(object.{{$entity.IdProperty.Name}}, 10, 64)
	if parseErr != nil {
		return parseErr
	}

	return box.Box.Remove(idUint64)
	{{- else -}}
	return box.Box.Remove(object.{{$entity.IdProperty.Name}})
	{{- end}}
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
// Note: this function panics if you try to create illegal queries; e.g. use properties of an alien type.
// This is typically a programming error. Use QueryOrError instead if you want the explicit error check.
func (box *{{$entity.Name}}Box) Query(conditions ...objectbox.Condition) *{{$entity.Name}}Query {
	return &{{$entity.Name}}Query{
		box.Box.Query(conditions...),
	}
}

// Creates a query with the given conditions. Use the fields of the {{$entity.Name}}_ struct to create conditions.
// Keep the *{{$entity.Name}}Query if you intend to execute the query multiple times.
func (box *{{$entity.Name}}Box) QueryOrError(conditions ...objectbox.Condition) (*{{$entity.Name}}Query, error) {
	if query, err := box.Box.QueryOrError(conditions...); err != nil {
		return nil, err
	} else {
		return &{{$entity.Name}}Query{query}, nil
	}
}

// Query provides a way to search stored objects
//
// For example, you can find all {{$entity.Name}} which {{$entity.IdProperty.Name}} is either 42 or 47:
// 		box.Query({{$entity.Name}}_.{{$entity.IdProperty.Name}}.In(42, 47)).Find()
type {{$entity.Name}}Query struct {
	*objectbox.Query
}

// Find returns all objects matching the query
func (query *{{$entity.Name}}Query) Find() ([]*{{$entity.Name}}, error) {
	objects, err := query.Query.Find()
	if err != nil {
		return nil, err
	}
	return objects.([]*{{$entity.Name}}), nil
}

{{end -}}`))
