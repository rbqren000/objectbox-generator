/*
 * Copyright (C) 2020 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

package templates

import (
	"text/template"
)

// TODO how to handle null values?
// TODO check failed allocs?

// CBindingTemplate is used to generated the binding code
var CBindingTemplate = template.Must(template.New("binding").Funcs(funcMap).Parse(
	`// Code generated by ObjectBox; DO NOT EDIT.

#pragma once

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#include "flatcc/flatcc.h"
#include "flatcc/flatcc_builder.h"
#include "objectbox.h"
{{range $entity := .Model.EntitiesWithMeta}}
{{PrintComments 0 $entity.Comments}}typedef struct {{$entity.Meta.CName}} {
	{{range $property := $entity.Properties}}{{$propType := PropTypeName $property.Type -}}
	{{PrintComments 1 $property.Comments}}{{if $property.Meta.FbIsVector}}{{$property.Meta.CElementType}}* {{$property.Meta.CppName}};
	{{- if or (eq $propType "StringVector") (eq $propType "ByteVector")}}
	size_t {{$property.Meta.CppName}}_len;{{end}}
	{{else}}{{$property.Meta.CppType}} {{$property.Meta.CppName}};
	{{end}}{{end}}
} {{$entity.Meta.CName}};

enum {{$entity.Meta.CName}}_ {
	{{$entity.Meta.CName}}_ENTITY_ID = {{$entity.Id.GetId}},
{{- range $property := $entity.Properties}}
	{{$entity.Meta.CName}}_PROP_ID_{{$property.Meta.CppName}} = {{$property.Id.GetId}},
{{- end}}
};
{{end}}
{{- range $entity := .Model.EntitiesWithMeta}}
/// Write given object to the FlatBufferBuilder
/// TODO test on a big-endian platform... especially vector creation
static bool {{$entity.Meta.CName}}_to_flatbuffer(flatcc_builder_t* B, const {{$entity.Meta.CName}}* object, void** out_buffer, size_t* out_size) {
    assert(B);
    assert(object);
    assert(out_buffer);
    assert(out_size);

    flatcc_builder_reset(B);
	flatcc_builder_start_buffer(B, 0, 0, 0);
	{{range $property := $entity.Properties}}{{$propType := PropTypeName $property.Type}}
	{{- if eq $propType "String"}}
	flatcc_builder_ref_t offset_{{$property.Meta.CppName}} = !object->{{$property.Meta.CppName}} ? 0 : flatcc_builder_create_string_str(B, object->{{$property.Meta.CppName}});
	{{- else if eq $propType "ByteVector"}}
	flatcc_builder_ref_t offset_{{$property.Meta.CppName}} = !object->{{$property.Meta.CppName}} ? 0 : flatcc_builder_create_vector(B, object->{{$property.Meta.CppName}}, object->{{$property.Meta.CppName}}_len, sizeof({{$property.Meta.CElementType}}), sizeof({{$property.Meta.CElementType}}), FLATBUFFERS_COUNT_MAX(sizeof({{$property.Meta.CElementType}})));
	{{- else if eq $propType "StringVector"}}
	flatcc_builder_ref_t offset_{{$property.Meta.CppName}} = 0;
	if (object->{{$property.Meta.CppName}}) {
		flatcc_builder_start_offset_vector(B);
		for (size_t i = 0; i < object->{{$property.Meta.CppName}}_len; i++) {
			flatcc_builder_ref_t ref = !object->{{$property.Meta.CppName}}[i] ? 0 : flatcc_builder_create_string_str(B, object->{{$property.Meta.CppName}}[i]);
			if (ref) flatcc_builder_offset_vector_push(B, ref);
		}
		offset_{{$property.Meta.CppName}} = flatcc_builder_end_offset_vector(B);
	}
	{{- end}}{{end}}

    if (flatcc_builder_start_table(B, {{len $entity.Properties}}) != 0) return false;

    void* p;
	flatcc_builder_ref_t* _p;
	{{range $property := $entity.Properties}}
	{{- if $property.Meta.FbIsVector}}
	if (offset_{{$property.Meta.CppName}}) {
        if (!(_p = flatcc_builder_table_add_offset(B, {{$property.FbSlot}}))) return false;
        *_p = offset_{{$property.Meta.CppName}};
    }
	{{- else}}
	if (!(p = flatcc_builder_table_add(B, {{$property.FbSlot}}, {{$property.Meta.FbTypeSize}}, {{$property.Meta.FbTypeSize}}))) return false;
    {{$property.Meta.FlatccFnPrefix}}_write_to_pe(p, object->{{$property.Meta.CppName}});
	{{- end}}
	{{end}}
    flatcc_builder_ref_t ref;
	if (!(ref = flatcc_builder_end_table(B))) return false;
	if (!flatcc_builder_end_buffer(B, ref)) return false;
    return (*out_buffer = flatcc_builder_finalize_aligned_buffer(B, out_size)) != NULL;
}

/// Read an object from a valid FlatBuffer.
/// If the read object contains vectors or strings, those are allocated on heap and must be freed after use by calling {{$entity.Meta.CName}}_free_pointers().
/// If the given object already contains un-freed pointers, the memory will be lost - free manually before calling this function twice on the same object. 
static void {{$entity.Meta.CName}}_from_flatbuffer(const void* data, size_t size, {{$entity.Meta.CName}}* out_object) {
	assert(data);
	assert(out_object);

	const uint8_t* table = (const uint8_t*) data + __flatbuffers_uoffset_read_from_pe(data);
	assert(table);
	flatbuffers_voffset_t *vt = (flatbuffers_voffset_t*) (table - __flatbuffers_soffset_read_from_pe(table));
	flatbuffers_voffset_t vs = __flatbuffers_voffset_read_from_pe(vt);

	// variables reused when reading strings and vectors
	flatbuffers_voffset_t offset;
	const flatbuffers_uoffset_t* val;
	size_t len;

	{{range $property := $entity.Properties}}{{$propType := PropTypeName $property.Type -}}
	{{if $property.Meta.FbIsVector}}
	if ((offset = (vs < sizeof(vt[0]) * ({{$property.FbSlot}} + 3)) ? {{$property.Meta.FbDefaultValue}} : __flatbuffers_voffset_read_from_pe(vt + {{$property.FbSlot}} + 2))) {
		val = (const flatbuffers_uoffset_t*)(table + offset + sizeof(flatbuffers_uoffset_t) + __flatbuffers_uoffset_read_from_pe(table + offset));
		len = (size_t) __flatbuffers_uoffset_read_from_pe(val - 1);
		out_object->{{$property.Meta.CppName}} = ({{$property.Meta.CElementType}}*) malloc({{if eq $propType "String"}}(len+1){{else}}len{{end}} * sizeof({{$property.Meta.CElementType}}));
		{{- if not (eq $propType "String")}}
		out_object->{{$property.Meta.CppName}}_len = len;
		{{- end -}}
		{{/*Note: direct copy for string and byte vectors*/}}
		{{if eq $propType "String"}}memcpy((void*)out_object->{{$property.Meta.CppName}}, (const void*)val, len+1);
		{{else if eq $propType "ByteVector"}}memcpy((void*)out_object->{{$property.Meta.CppName}}, (const void*)val, len);
		{{else}}{{/* StringVector - FB vector contains offsets to strings, each must be read separately*/ -}}
		for (size_t i = 0; i < len; i++, val++) {
			const uint8_t* str = (const uint8_t*) val + (size_t)__flatbuffers_uoffset_read_from_pe(val) + sizeof(val[0]);
			out_object->{{$property.Meta.CppName}}[i] = (char*) malloc((strlen((const char*)str) + 1) * sizeof(char));
			strcpy((char*)out_object->{{$property.Meta.CppName}}[i], (const char*)str);
		}{{end}}
	} else {
		out_object->{{$property.Meta.CppName}} = NULL;
		{{- if not (eq $propType "String")}}
		out_object->{{$property.Meta.CppName}}_len = 0;
		{{- end}}
	}
	{{else}}out_object->{{$property.Meta.CppName}} = (vs < sizeof(vt[0]) * ({{$property.FbSlot}} + 3)) ? {{$property.Meta.FbDefaultValue}} : {{$property.Meta.FlatccFnPrefix}}_read_from_pe(table + __flatbuffers_voffset_read_from_pe(vt + {{$property.FbSlot}} + 2));
	{{- end}}
	{{end}}
}

/// Read an object from a valid FlatBuffer, allocating the object on heap. 
/// The object must be freed after use by calling {{$entity.Meta.CName}}_free();
static {{$entity.Meta.CName}}* {{$entity.Meta.CName}}_new_from_flatbuffer(const void* data, size_t size) {
	{{$entity.Meta.CName}}* object = ({{$entity.Meta.CName}}*) malloc(sizeof({{$entity.Meta.CName}}));
	{{$entity.Meta.CName}}_from_flatbuffer(data, size, object);
	return object;
}

static void {{$entity.Meta.CName}}_free_pointers({{$entity.Meta.CName}}* object) {
	{{- range $property := $entity.Properties}}{{$propType := PropTypeName $property.Type}}{{if $property.Meta.FbIsVector}}
	if (object->{{$property.Meta.CppName}}) {
		{{- if eq $propType "StringVector"}}
		assert(object->{{$property.Meta.CppName}}_len > 0);
		for (size_t i = 0; i < object->{{$property.Meta.CppName}}_len; i++) {
			if (object->{{$property.Meta.CppName}}[i]) free(object->{{$property.Meta.CppName}}[i]);
		}{{end}}
		free(object->{{$property.Meta.CppName}});
		object->{{$property.Meta.CppName}} = NULL;
	{{- if not (eq $propType "String")}}
		object->{{$property.Meta.CppName}}_len = 0;
	} else {
		assert(object->{{$property.Meta.CppName}}_len == 0);
	{{- end}}
	}
	{{end}}
	{{- end}}
}

static void {{$entity.Meta.CName}}_free({{$entity.Meta.CName}}** object) {
	if (!object || !*object) return;
	{{$entity.Meta.CName}}_free_pointers(*object);
	free(*object);
	*object = NULL;
}
{{end}}
`))
