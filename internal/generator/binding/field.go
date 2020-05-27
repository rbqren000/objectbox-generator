/*
 * Copyright 2019 ObjectBox Ltd. All rights reserved.
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

package binding

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/objectbox/objectbox-go/internal/generator/go"
	"github.com/objectbox/objectbox-go/internal/generator/model"
)

// Field holds common field/property information used by specialized code parsers/generators.
// Additionally, it groups some shared logic, e.g. annotation processing
type Field struct {
	Name      string
	IsSkipped bool

	property *model.Property
}

func CreateField(prop *model.Property) *Field {
	return &Field{property: prop}
}

func (field *Field) SetName(name string) {
	field.Name = name
	if len(field.property.Name) == 0 {
		field.property.Name = strings.ToLower(name)
	}
}

// ProcessAnnotations checks all set annotations for any inconsistencies and sets local/property fields (flags, name, ...)
// TODO move generator.Annotation to this package
func (field *Field) ProcessAnnotations(a map[string]*gogenerator.Annotation) error {
	if a["-"] != nil {
		if len(a) != 1 || a["-"].Value != "" {
			return errors.New("to ignore the property, use only `objectbox:\"-\"` as an annotation")
		}
		field.IsSkipped = true
		return nil
	}

	if a["id"] != nil {
		field.property.AddFlag(model.PropertyFlagId)
	}

	if a["name"] != nil {
		if len(a["name"].Value) == 0 {
			return fmt.Errorf("name annotation value must not be empty - it's the field name in DB")
		}
		field.property.Name = strings.ToLower(a["name"].Value)
	}

	if a["date"] != nil {
		if field.property.Type != model.PropertyTypeLong {
			return fmt.Errorf("invalid underlying type (PropertyType %v) for date field; expecting long", field.property.Type)
		}
		field.property.Type = model.PropertyTypeDate
	}

	if a["index"] != nil {
		switch strings.ToLower(a["index"].Value) {
		case "":
			// if the user doesn't define index type use the default based on the data-type
			if field.property.Type == model.PropertyTypeString {
				field.property.AddFlag(model.PropertyFlagIndexHash)
			} else {
				field.property.AddFlag(model.PropertyFlagIndexed)
			}
		case "value":
			field.property.AddFlag(model.PropertyFlagIndexed)
		case "hash":
			field.property.AddFlag(model.PropertyFlagIndexHash)
		case "hash64":
			field.property.AddFlag(model.PropertyFlagIndexHash64)
		default:
			return fmt.Errorf("unknown index type %s", a["index"].Value)
		}

		if err := field.property.SetIndex(); err != nil {
			return err
		}
	}

	if a["unique"] != nil {
		field.property.AddFlag(model.PropertyFlagUnique)

		if err := field.property.SetIndex(); err != nil {
			return err
		}
	}

	if a["uid"] != nil {
		if len(a["uid"].Value) == 0 {
			// in case the user doesn't provide `objectbox:"uid"` value, it's considered in-process of setting up UID
			// this flag is handled by the merge mechanism and prints the UID of the already existing property
			field.property.UidRequest = true
		} else if uid, err := strconv.ParseUint(a["uid"].Value, 10, 64); err != nil {
			return fmt.Errorf("can't parse uid - %s", err)
		} else if id, err := field.property.Id.GetIdAllowZero(); err != nil {
			return fmt.Errorf("can't parse property Id - %s", err)
		} else {
			field.property.Id = model.CreateIdUid(id, uid)
		}
	}

	return nil
}