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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/objectbox/objectbox-generator/internal/generator"
	"github.com/objectbox/objectbox-generator/internal/generator/c"
	"github.com/objectbox/objectbox-generator/internal/generator/flatbuffersc"
)

const defaultErrorCode = 2

func main() {
	if runFlatcIfRequested() {
		return
	}

	path, clean, options := getArgs()

	var err error
	if clean {
		fmt.Printf("Removing ObjectBox bindings for %s\n", path)
		err = generator.Clean(options.CodeGenerator, path)
	} else {
		fmt.Printf("Generating ObjectBox bindings for %s\n", path)
		err = generator.Process(path, options)
	}

	stopOnError(0, err)
}

func stopOnError(code int, err error) {
	if err != nil {
		fmt.Println(err)
		if code == 0 {
			code = defaultErrorCode
		}
		os.Exit(code)
	}
}

func showUsage() {
	fmt.Fprint(flag.CommandLine.Output(), `Usage:
  objectbox-generator [flags] [path-pattern]
      to generate the binding code

or
  objectbox-generator clean [path-pattern]
      to remove the generated files instead of creating them - this removes *.obx.h and objectbox-model.h but keeps objectbox-model.json

or
  objectbox-generator FLATC [flatc arguments]
      to execute FlatBuffers flatc command line tool Any arguments after the FLATC keyword are passed through.

path-pattern:
  * a path or a valid path pattern (e.g. ./...)
  
Available flags:
`)
	flag.PrintDefaults()
}

func showUsageAndExit(a ...interface{}) {
	if len(a) > 0 {
		a = append(a, "\n\n")
		fmt.Fprint(flag.CommandLine.Output(), a...)
	}
	showUsage()
	os.Exit(1)
}

func getArgs() (path string, clean bool, options generator.Options) {
	var gen = &cgenerator.CGenerator{}
	options.CodeGenerator = gen

	var outputs = make(map[string]*bool)

	var printVersion bool
	var printHelp bool
	flag.Usage = showUsage
	outputs["c"] = flag.Bool("c", false, "generate plain C code")
	outputs["cpp"] = flag.Bool("cpp", false, "generate C++ code ")
	flag.StringVar(&gen.OutPath, "out", "", "output path for generated source files")
	flag.StringVar(&options.ModelInfoFile, "persist", "", "path to the model information persistence file (JSON)")
	flag.BoolVar(&printVersion, "version", false, "print the generator version info")
	flag.BoolVar(&printHelp, "help", false, "print this help")
	flag.Parse()

	if printHelp {
		showUsage()
		os.Exit(0)
	}

	if printVersion {
		fmt.Println(fmt.Sprintf("ObjectBox code generator version: %d", generator.Version))
		os.Exit(0)
	}

	if flag.NArg() == 2 {
		clean = true
		if flag.Arg(0) != "clean" {
			showUsageAndExit("Unknown argument %s\n", flag.Arg(0))
		}

		path = flag.Arg(1)

	} else if flag.NArg() == 1 {
		path = flag.Arg(0)
	} else if flag.NArg() != 0 {
		showUsageAndExit()
	}

	if len(path) == 0 {
		showUsageAndExit()
	}

	if *outputs["cpp"] && *outputs["c"] {
		showUsageAndExit("Only one output language can be specified at the moment")
	} else if *outputs["c"] {
		gen.PlainC = true
	} else if !*outputs["cpp"] {
		showUsageAndExit("You must specify an output language")
	}

	return
}

// runFlatcIfRequested checks command line arguments and if they start with FLATC, executes flatc compiler with the remainder of the arguments
func runFlatcIfRequested() bool {
	if len(os.Args) < 2 || strings.ToLower(os.Args[1]) != "flatc" {
		return false
	}

	stopOnError(flatbuffersc.ExecuteFlatc(os.Args[2:]))
	return true
}