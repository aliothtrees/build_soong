// Copyright 2018 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tradefed

import (
	"fmt"
	"strings"

	"github.com/google/blueprint"

	"android/soong/android"
)

func getTestConfigTemplate(ctx android.ModuleContext, prop *string) android.OptionalPath {
	return ctx.ExpandOptionalSource(prop, "test_config_template")
}

func getTestConfig(ctx android.ModuleContext, prop *string) android.Path {
	if p := ctx.ExpandOptionalSource(prop, "test_config"); p.Valid() {
		return p.Path()
	} else if p := android.ExistentPathForSource(ctx, ctx.ModuleDir(), "AndroidTest.xml"); p.Valid() {
		return p.Path()
	}
	return nil
}

var autogenTestConfig = pctx.StaticRule("autogenTestConfig", blueprint.RuleParams{
	Command: "sed 's&{MODULE}&${name}&g' $template > $out &&" +
		"${optionCmd} $out",
	CommandDeps: []string{"$template"},
}, "name", "template", "optionCmd")

func testConfigPath(ctx android.ModuleContext, prop *string) (path android.Path, autogenPath android.WritablePath) {
	if p := getTestConfig(ctx, prop); p != nil {
		return p, nil
	} else if !strings.HasPrefix(ctx.ModuleDir(), "cts/") {
		outputFile := android.PathForModuleOut(ctx, ctx.ModuleName()+".config")
		return nil, outputFile
	} else {
		// CTS modules can be used for test data, so test config files must be
		// explicitly created using AndroidTest.xml
		// TODO(b/112602712): remove the path check
		return nil, nil
	}
}

func autogenTemplate(ctx android.ModuleContext, output android.WritablePath, template string, optionsMap map[string]string) {
	// If no test option found, delete {UID_OPTION} line.
	// If found, replace it with corresponding options format.
	optionCmd := "sed -i '/{UID_OPTION}/d'"
	if optionsMap != nil {
		//Append options
		var options []string
		for optionName, value := range optionsMap {
			if value != "" {
				options = append(options, fmt.Sprintf("<option name=\"%s\" value=\"%s\" />", optionName, value))
			}
		}
		optionCmd = fmt.Sprintf("sed -i 's&{UID_OPTION}&%s&g'", strings.Join(options, "\\n        "))
	}
	ctx.Build(pctx, android.BuildParams{
		Rule:        autogenTestConfig,
		Description: "test config",
		Output:      output,
		Args: map[string]string{
			"name":      ctx.ModuleName(),
			"template":  template,
			"optionCmd": optionCmd,
		},
	})
}

func AutoGenNativeTestConfig(ctx android.ModuleContext, testConfigProp *string,
	testConfigTemplateProp *string, optionsMap map[string]string) android.Path {
	path, autogenPath := testConfigPath(ctx, testConfigProp)
	if autogenPath != nil {
		templatePath := getTestConfigTemplate(ctx, testConfigTemplateProp)
		if templatePath.Valid() {
			autogenTemplate(ctx, autogenPath, templatePath.String(), optionsMap)
		} else {
			if ctx.Device() {
				autogenTemplate(ctx, autogenPath, "${NativeTestConfigTemplate}", optionsMap)
			} else {
				autogenTemplate(ctx, autogenPath, "${NativeHostTestConfigTemplate}", optionsMap)
			}
		}
		return autogenPath
	}
	return path
}

func AutoGenNativeBenchmarkTestConfig(ctx android.ModuleContext, testConfigProp *string,
	testConfigTemplateProp *string) android.Path {
	path, autogenPath := testConfigPath(ctx, testConfigProp)
	if autogenPath != nil {
		templatePath := getTestConfigTemplate(ctx, testConfigTemplateProp)
		if templatePath.Valid() {
			autogenTemplate(ctx, autogenPath, templatePath.String(), nil)
		} else {
			autogenTemplate(ctx, autogenPath, "${NativeBenchmarkTestConfigTemplate}", nil)
		}
		return autogenPath
	}
	return path
}

func AutoGenJavaTestConfig(ctx android.ModuleContext, testConfigProp *string, testConfigTemplateProp *string) android.Path {
	path, autogenPath := testConfigPath(ctx, testConfigProp)
	if autogenPath != nil {
		templatePath := getTestConfigTemplate(ctx, testConfigTemplateProp)
		if templatePath.Valid() {
			autogenTemplate(ctx, autogenPath, templatePath.String(), nil)
		} else {
			if ctx.Device() {
				autogenTemplate(ctx, autogenPath, "${JavaTestConfigTemplate}", nil)
			} else {
				autogenTemplate(ctx, autogenPath, "${JavaHostTestConfigTemplate}", nil)
			}
		}
		return autogenPath
	}
	return path
}

var autogenInstrumentationTest = pctx.StaticRule("autogenInstrumentationTest", blueprint.RuleParams{
	Command: "${AutoGenTestConfigScript} $out $in ${EmptyTestConfig} $template",
	CommandDeps: []string{
		"${AutoGenTestConfigScript}",
		"${EmptyTestConfig}",
		"$template",
	},
}, "name", "template")

func AutoGenInstrumentationTestConfig(ctx android.ModuleContext, testConfigProp *string, testConfigTemplateProp *string, manifest android.Path) android.Path {
	path, autogenPath := testConfigPath(ctx, testConfigProp)
	if autogenPath != nil {
		template := "${InstrumentationTestConfigTemplate}"
		moduleTemplate := getTestConfigTemplate(ctx, testConfigTemplateProp)
		if moduleTemplate.Valid() {
			template = moduleTemplate.String()
		}
		ctx.Build(pctx, android.BuildParams{
			Rule:        autogenInstrumentationTest,
			Description: "test config",
			Input:       manifest,
			Output:      autogenPath,
			Args: map[string]string{
				"name":     ctx.ModuleName(),
				"template": template,
			},
		})
		return autogenPath
	}
	return path
}
