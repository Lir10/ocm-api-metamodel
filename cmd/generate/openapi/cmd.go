/*
Copyright (c) 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package openapi

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/openshift-online/ocm-api-metamodel/pkg/generators"
	"github.com/openshift-online/ocm-api-metamodel/pkg/generators/openapi"
	"github.com/openshift-online/ocm-api-metamodel/pkg/http"
	"github.com/openshift-online/ocm-api-metamodel/pkg/language"
	"github.com/openshift-online/ocm-api-metamodel/pkg/reporter"
)

// Cmd is the definition of the command:
var Cmd = &cobra.Command{
	Use:   "openapi",
	Short: "Generate OpenAPI specifications",
	Long:  "Generate OpenAPI specifications.",
	Run:   run,
}

// Values of the command line arguments:
var args struct {
	paths  []string
	output string
}

func init() {
	flags := Cmd.Flags()
	flags.StringSliceVar(
		&args.paths,
		"model",
		[]string{},
		"File or directory containing the model. If it is a directory then all .model"+
			"files inside it and its sub directories will be loaded. If used "+
			"multiple times then all the specified files and directories will be "+
			"loaded, in the same order that they appear in the command line.",
	)
	flags.StringVar(
		&args.output,
		"output",
		"",
		"Directory where the OpenAPI specifications will be generated.",
	)
}

func run(cmd *cobra.Command, argv []string) {
	// Create the reporter:
	reporter := reporter.NewReporter()

	// Check command line options:
	ok := true
	if len(args.paths) == 0 {
		reporter.Errorf("Option '--model' is mandatory")
		ok = false
	}
	if args.output == "" {
		reporter.Errorf("Option '--output' is mandatory")
		ok = false
	}
	if !ok {
		os.Exit(1)
	}

	// Read the model:
	model, err := language.NewReader().
		Reporter(reporter).
		Inputs(args.paths).
		Read()
	if err != nil {
		reporter.Errorf("Can't read model: %v", err)
		os.Exit(1)
	}

	// Create the calculators:
	bindingCalculator, err := http.NewBindingCalculator().
		Reporter(reporter).
		Build()
	if err != nil {
		reporter.Errorf("Can't create HTTP binding calculator: %v", err)
		os.Exit(1)
	}
	openapiCalculator, err := openapi.NewNamesCalculator().
		Reporter(reporter).
		Build()
	if err != nil {
		reporter.Errorf("Can't create OpenAPI names calculator: %v", err)
		os.Exit(1)
	}

	// We will store here all the code generators that we will later run:
	var gens []generators.Generator
	var gen generators.Generator

	// Create the OpenAPI specifications generator:
	gen, err = openapi.NewOpenAPIGenerator().
		Reporter(reporter).
		Model(model).
		Output(args.output).
		Names(openapiCalculator).
		Binding(bindingCalculator).
		Build()
	if err != nil {
		reporter.Errorf("Can't create OpenAPI generator: %v", err)
		os.Exit(1)
	}
	gens = append(gens, gen)

	// Run the generators:
	for _, gen := range gens {
		err = gen.Run()
		if err != nil {
			reporter.Errorf("Generation failed: %v", err)
			os.Exit(1)
		}
	}

	// Bye:
	os.Exit(0)
}
