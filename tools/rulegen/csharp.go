package main

var csharpLibraryWorkspaceTemplateString = `load("@rules_proto_grpc//{{ .Lang.Dir }}:repositories.bzl", rules_proto_grpc_{{ .Lang.Name }}_repos="{{ .Lang.Name }}_repos")

rules_proto_grpc_{{ .Lang.Name }}_repos()

load(
    "@io_bazel_rules_dotnet//dotnet:defs.bzl",
    "core_register_sdk",
    "dotnet_register_toolchains",
    "dotnet_repositories",
)

core_version = "v2.1.503"

dotnet_register_toolchains(
    core_version = core_version,
)

dotnet_register_toolchains(
    core_version = core_version,
)

core_register_sdk(
    name = "core_sdk",
    core_version = core_version,
)

dotnet_repositories()

load("@rules_proto_grpc//csharp/nuget:packages.bzl", nuget_packages = "packages")

nuget_packages()

load("@rules_proto_grpc//csharp/nuget:nuget.bzl", "nuget_protobuf_packages")

nuget_protobuf_packages()`

var csharpProtoLibraryWorkspaceTemplate = mustTemplate(csharpLibraryWorkspaceTemplateString)

var csharpGrpcLibraryWorkspaceTemplate = mustTemplate(csharpLibraryWorkspaceTemplateString + `

load("@rules_proto_grpc//csharp/nuget:nuget.bzl", "nuget_grpc_packages")

nuget_grpc_packages()`)

var csharpLibraryRuleTemplateString = `load("//{{ .Lang.Dir }}:{{ .Lang.Name }}_{{ .Rule.Kind }}_compile.bzl", "{{ .Lang.Name }}_{{ .Rule.Kind }}_compile")
load("@io_bazel_rules_dotnet//dotnet:defs.bzl", "core_library")

def {{ .Rule.Name }}(**kwargs):
    # Compile protos
    name_pb = kwargs.get("name") + "_pb"
    {{ .Lang.Name }}_{{ .Rule.Kind }}_compile(
        name = name_pb,
        **{k: v for (k, v) in kwargs.items() if k in ("deps", "verbose")} # Forward args
    )
`

var csharpProtoLibraryRuleTemplate = mustTemplate(csharpLibraryRuleTemplateString + `
    # Create {{ .Lang.Name }} library
    core_library(
        name = kwargs.get("name"),
        srcs = [name_pb],
        deps = PROTO_DEPS,
        visibility = kwargs.get("visibility"),
    )

PROTO_DEPS = [
    "@google.protobuf//:netstandard1.0_core",
    "@io_bazel_rules_dotnet//dotnet/stdlib.core:system.io.dll",
]`)

var csharpGrpcLibraryRuleTemplate = mustTemplate(csharpLibraryRuleTemplateString + `
    # Create {{ .Lang.Name }} library
    core_library(
        name = kwargs.get("name"),
        srcs = [name_pb],
        deps = GRPC_DEPS,
        visibility = kwargs.get("visibility"),
    )

GRPC_DEPS = [
    "@google.protobuf//:netstandard1.0_core",
    "@io_bazel_rules_dotnet//dotnet/stdlib.core:system.io.dll",
    "@grpc.core//:netstandard1.5_core",
    "@system.interactive.async//:netstandard2.0_core",
]`)

var csharpLibraryFlags = []*Flag{
	{
		Category:    "build",
		Name:        "strategy",
		Value:       "CoreCompile=standalone",
		Description: "dotnet SDK desperately wants to find the HOME directory",
	},
}

func makeCsharp() *Language {
	return &Language{
		Dir:   "csharp",
		Name:  "csharp",
		DisplayName: "C#",
		Flags: commonLangFlags,
		SkipTestPlatforms: []string{"all"},
		Notes: mustTemplate(`Rules for generating C# protobuf and gRPC ` + "`.cs`" + ` files and libraries using standard Protocol Buffers and gRPC. Libraries are created with ` + "`core_library`" + ` from [rules_dotnet](https://github.com/bazelbuild/rules_dotnet)

**NOTE 1**: the csharp_* rules currently don't play nicely with sandboxing.  You may see errors like:

~~~python
The user's home directory could not be determined. Set the 'DOTNET_CLI_HOME' environment variable to specify the directory to use.
~~~

or

~~~python
System.ArgumentNullException: Value cannot be null.
Parameter name: path1
   at System.IO.Path.Combine(String path1, String path2)
   at Microsoft.DotNet.Configurer.CliFallbackFolderPathCalculator.get_DotnetUserProfileFolderPath()
   at Microsoft.DotNet.Configurer.FirstTimeUseNoticeSentinel..ctor(CliFallbackFolderPathCalculator cliFallbackFolderPathCalculator)
   at Microsoft.DotNet.Cli.Program.ProcessArgs(String[] args, ITelemetry telemetryClient)
   at Microsoft.DotNet.Cli.Program.Main(String[] args)
~~~

To remedy this, use --strategy=CoreCompile=standalone for the csharp rules (put it in your .bazelrc file).

**NOTE 2**: the csharp nuget dependency sha256 values do not appear stable.`),
		Rules: []*Rule{
			&Rule{
				Name:             "csharp_proto_compile",
				Kind:             "proto",
				Implementation:   aspectRuleTemplate,
				Plugins:          []string{"//csharp:csharp_plugin"},
				WorkspaceExample: protoWorkspaceTemplate,
				BuildExample:     protoCompileExampleTemplate,
				Doc:              "Generates C# protobuf `.cs` artifacts",
				Attrs:            aspectProtoCompileAttrs,
				SkipTestPlatforms: []string{"none"},
			},
			&Rule{
				Name:             "csharp_grpc_compile",
				Kind:             "grpc",
				Implementation:   aspectRuleTemplate,
				Plugins:          []string{"//csharp:csharp_plugin", "//csharp:grpc_csharp_plugin"},
				WorkspaceExample: grpcWorkspaceTemplate,
				BuildExample:     grpcCompileExampleTemplate,
				Doc:              "Generates C# protobuf+gRPC `.cs` artifacts",
				Attrs:            aspectProtoCompileAttrs,
				SkipTestPlatforms: []string{"none"},
			},
			&Rule{
				Name:             "csharp_proto_library",
				Kind:             "proto",
				Implementation:   csharpProtoLibraryRuleTemplate,
				WorkspaceExample: csharpProtoLibraryWorkspaceTemplate,
				BuildExample:     protoLibraryExampleTemplate,
				Doc:              "Generates a C# protobuf library using `core_library` from `rules_dotnet`",
				Attrs:            aspectProtoCompileAttrs,
				Flags:            csharpLibraryFlags,
				Experimental:     true, // Due to failing dependencies
			},
			&Rule{
				Name:             "csharp_grpc_library",
				Kind:             "grpc",
				Implementation:   csharpGrpcLibraryRuleTemplate,
				WorkspaceExample: csharpGrpcLibraryWorkspaceTemplate,
				BuildExample:     grpcLibraryExampleTemplate,
				Doc:              "Generates a C# protobuf+gRPC library using `core_library` from `rules_dotnet`",
				Attrs:            aspectProtoCompileAttrs,
				Flags:            csharpLibraryFlags,
				Experimental:     true, // Due to failing dependencies
			},
		},
	}
}
