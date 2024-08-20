/*
 * ObjectBox Generator - a build time tool for ObjectBox
 * Copyright (C) 2018-2024 ObjectBox Ltd. All rights reserved.
 * https://objectbox.io
 *
 * This file is part of ObjectBox Generator.
 *
 * ObjectBox Generator is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * ObjectBox Generator is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with ObjectBox Generator.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package generator provides tools to generate ObjectBox entity bindings between GO structs & ObjectBox schema
package generator

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/objectbox/objectbox-generator/v4/internal/generator/model"
)

// Version specifies the current generator version.
const Version = "4.0.0-alpha3"

// VersionId specifies the current generator version identifier.
// It is used to validate generated code compatibility and is increased when there are changes in the generated code.
// This validation seems to be limited to Go: the generated code "knows" its version,
// and thus we can check at runtime if the generated code matches the lib version.
// Internal generator changes that don't change the output (in an incompatible way) do not cause an increase.
const VersionId = 6

// ModelInfoFile returns the model info JSON file name in the given directory
func ModelInfoFile(dir string) string {
	return filepath.Join(dir, "objectbox-model.json")
}

// CodeGenerator interface is used to abstract per-language generators, e.g. for Go, C, C++, etc
type CodeGenerator interface {
	// BindingFiles returns the names of language binding files for the given entity file.
	// TODO "binding files" is not intuitive name (especially without mentioning "**language** binding").
	//      Rename to "language", "generated" or "output" files instead?
	//      Rename functions, variables, etc. accordingly.
	BindingFiles(forFile string, options Options) []string

	// ModelFile returns the language-specific model source file for the given JSON info file path
	ModelFile(forFile string, options Options) string

	// IsGeneratedFile returns true if the given path is recognized as a file generated by this generator
	IsGeneratedFile(file string) bool

	// IsSourceFile returns true if the given path is recognized as an input file by this generator.
	// E.g. for Go files, ending with ".go", and for C++ ending with ".fbs".
	IsSourceFile(file string) bool

	// ParseSource reads the input file and creates a model representation
	ParseSource(sourceFile string) (*model.ModelInfo, error)

	// WriteBindingFiles generates and writes binding source code files
	WriteBindingFiles(sourceFile string, options Options, mergedModel *model.ModelInfo) error

	// WriteModelBindingFile generates and writes binding source code file for model setup
	WriteModelBindingFile(options Options, mergedModel *model.ModelInfo) error
}

// WriteFile writes data to targetFile, while using permissions either from the targetFile or permSource
func WriteFile(file string, data []byte, permSource string) error {
	var perm os.FileMode
	// copy permissions either from the existing file or from the source file
	if info, _ := os.Stat(file); info != nil {
		perm = info.Mode()
	} else if info, err := os.Stat(permSource); info != nil {
		perm = info.Mode()
	} else {
		return err
	}

	return ioutil.WriteFile(file, data, perm)
}

// Process is the main API method of the package
// it takes source file & model-information file paths and generates bindings (as a sibling file to the source file)
func Process(options Options) error {
	var err error

	// Ensure output directory is existing or create
	if len(options.OutPath) != 0 {
		err := os.MkdirAll(options.OutPath, 0750)
		if err != nil {
			return fmt.Errorf("can't create output path '"+options.OutPath+"': %s", err)
		}
	}

	// Ensure output header directory is existing or create
	if len(options.OutHeadersPath) != 0 {
		err := os.MkdirAll(options.OutHeadersPath, 0750)
		if err != nil {
			return fmt.Errorf("can't create output headers path '"+options.OutPath+"': %s", err)
		}
	}

	if PathIsDirOrPattern(options.InPath) {
		var additional string
		var cleanPath = options.InPath
		if len(options.OutPath) != 0 {
			additional = "of output path (-out=" + options.OutPath + ") "
			cleanPath = options.OutPath
		}
		fmt.Printf("Requested to generate for directory/pattern %s, performing an implicit cleanup %sfirst\n", options.InPath, additional)
		err = Clean(options.CodeGenerator, cleanPath)
		if err != nil {
			return err
		}
	}

	// if no random generator is provided, we create and seed a new one
	if options.Rand == nil {
		options.Rand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	}

	if len(options.ModelInfoFile) == 0 {
		options.ModelInfoFile = ModelInfoFile(filepath.Dir(options.InPath))
	}

	var modelInfo *model.ModelInfo

	modelInfo, err = model.LoadOrCreateModel(options.ModelInfoFile)
	if err != nil {
		return fmt.Errorf("can't init ModelInfo: %s", err)
	}

	modelInfo.Rand = options.Rand
	defer modelInfo.Close()

	if err = modelInfo.Validate(); err != nil {
		return fmt.Errorf("invalid ModelInfo loaded: %s", err)
	}

	// if the model is valid, upgrade it to the latest version
	modelInfo.MinimumParserVersion = model.ModelVersion
	modelInfo.ModelVersion = model.ModelVersion

	if err = createBinding(options, modelInfo); err != nil {
		return err
	}

	if err = createModel(options, modelInfo); err != nil {
		return err
	}

	return nil
}

func createBinding(options Options, storedModel *model.ModelInfo) error {
	return pathForEach(options.InPath, func(filePath string) error {
		if !options.CodeGenerator.IsSourceFile(filePath) {
			return nil
		}

		// clear meta information from the previous createBinding() call (when processing multiple files at once)
		for _, entity := range storedModel.EntitiesWithMeta() {
			entity.Meta = nil
		}

		currentModel, err := options.CodeGenerator.ParseSource(filePath)
		if err != nil {
			return err
		}

		if err = mergeBindingWithModelInfo(currentModel, storedModel); err != nil {
			return fmt.Errorf("can't merge model information: %s", err)
		}

		if err = storedModel.Finalize(); err != nil {
			return fmt.Errorf("model finalization failed: %s", err)
		}

		if err = options.CodeGenerator.WriteBindingFiles(filePath, options, storedModel); err != nil {
			return err
		}

		for _, entity := range storedModel.EntitiesWithMeta() {
			entity.CurrentlyPresent = true
		}

		return nil
	})
}

func createModel(options Options, modelInfo *model.ModelInfo) error {
	// clean entities not present in the current run - ONLY if running for a path
	if PathIsDirOrPattern(options.InPath) {
		removedEntities := make([]*model.Entity, 0)
		for _, entity := range modelInfo.Entities {
			if !entity.CurrentlyPresent {
				fmt.Printf("Removing missing entity %s %s from the model\n", entity.Name, entity.Id)
				removedEntities = append(removedEntities, entity)
			}
		}

		for _, entity := range removedEntities {
			if err := modelInfo.RemoveEntity(entity); err != nil {
				return fmt.Errorf("removing entity %s failed: %s", entity.Name, err)
			}
		}

		if err := modelInfo.Finalize(); err != nil {
			return fmt.Errorf("model finalization failed: %s", err)
		}
	}

	if err := modelInfo.Write(); err != nil {
		return fmt.Errorf("can't write model-info file %s: %s", options.ModelInfoFile, err)
	}

	return options.CodeGenerator.WriteModelBindingFile(options, modelInfo)
}

// Clean removes generated files in the given path.
// Removes *.obx.* and objectbox-model.[go|h|...] but keeps objectbox-model.json
func Clean(codeGenerator CodeGenerator, path string) error {
	return pathForEach(path, func(filePath string) error {
		if !codeGenerator.IsGeneratedFile(filePath) {
			return nil
		}
		fmt.Printf("Removing %s\n", filePath)
		return os.Remove(filePath)
	})
}

const recursionSuffix = "/..."

// PathIsDirOrPattern checks whether the given path is a path pattern, a directory or a single file.
func PathIsDirOrPattern(path string) bool {
	// if it's a recursion pattern
	if strings.HasSuffix(path, recursionSuffix) {
		return true
	}

	// if it's a Glob pattern (see hasMeta() in package path/filepath/match.go)
	if strings.ContainsAny(path, `*?[`) || (runtime.GOOS != "windows" && strings.ContainsAny(path, `\`)) {
		return true
	}

	// if it's a directory
	if finfo, err := os.Stat(path); err == nil && finfo.IsDir() {
		return true
	}

	return false
}

// pathForEach executes the given function for each file in the given directory/path pattern
func pathForEach(path string, fn func(filePath string) error) error {
	var recursive bool

	// if it's a pattern
	if strings.HasSuffix(path, recursionSuffix) {
		recursive = true
		path = path[0:len(path)-len(recursionSuffix)] + "/*"
	} else {
		// if it's a directory
		if finfo, err := os.Stat(path); err == nil && finfo.IsDir() {
			path = path + "/*"
		}
	}

	matches, err := filepath.Glob(path)
	if err != nil {
		return err
	}

	for _, subpath := range matches {
		finfo, err := os.Stat(subpath)
		if err != nil {
			return err
		}

		if recursive && finfo.Mode().IsDir() {
			err = pathForEach(subpath+recursionSuffix, fn)
		} else if finfo.Mode().IsRegular() {
			err = fn(subpath)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
