## Example Usage

These steps walk through the steps to go from a raw `.proto` file to a C++
library:


**Step 1**: Write a Protocol Buffer file (example: `thing.proto`):

```proto
syntax = "proto3";

package example;

import "google/protobuf/any.proto";

message Thing {
    string name = 1;
    google.protobuf.Any payload = 2;
}
```


**Step 2**: Write a `BAZEL.build` file with a native [`proto_library`](https://docs.bazel.build/versions/master/be/protocol-buffer.html#proto_library)
rule:

```python
proto_library(
    name = "thing_proto",
    srcs = ["thing.proto"],
    deps = ["@com_google_protobuf//:any_proto"],
)
```

In this example we have a dependency on a well-known type `any.proto`, hence the
`proto_library` to `proto_library` dependency (`"@com_google_protobuf//:any_proto"`)


**Step 3**: Add a `cpp_proto_compile` rule (substitute `cpp_` for the language
of your choice).

> NOTE: In this example `thing.proto` does not include service definitions
(gRPC).  For protos with services, use the `cpp_grpc_compile` rule instead.

```python
# BUILD.bazel
load("@rules_proto_grpc//cpp:defs.bzl", "cpp_proto_compile")

cpp_proto_compile(
    name = "cpp_thing_proto",
    deps = [":thing_proto"],
)
```

But wait, before we can build this, we need to load the dependencies necessary
for this rule (from [cpp/README.md](/cpp/README.md)):


**Step 4**: Load the workspace macro corresponding to the build rule.

```python
# WORKSPACE
load("@rules_proto_grpc//cpp:repositories.bzl", "cpp_repos")

cpp_repos()
```

We're now ready to build the rule:


**Step 5**: Build it!

```sh
$ bazel build //example/proto:cpp_thing_proto
Target //example/proto:cpp_thing_proto up-to-date:
  bazel-genfiles/example/proto/cpp_thing_proto/example/proto/thing.pb.h
  bazel-genfiles/example/proto/cpp_thing_proto/example/proto/thing.pb.cc
```

If we were only interested in the generated file artifacts, the
`cpp_grpc_compile` rule would be fine. However, for convenience we'd rather
have the outputs compiled into an `*.so` file. To do that, let's change the
rule from `cpp_proto_compile` to `cpp_proto_library`:

```python
# BUILD.bazel
load("@rules_proto_grpc//cpp:defs.bzl", "cpp_proto_library")

cpp_proto_library(
    name = "cpp_thing_proto",
    deps = [":thing_proto"],
)
```

```sh
$ bazel build //example/proto:cpp_thing_proto
Target //example/proto:cpp_thing_proto up-to-date:
  bazel-bin/example/proto/libcpp_thing_proto.a
  bazel-bin/example/proto/libcpp_thing_proto.so  bazel-genfiles/example/proto/cpp_thing_proto/example/proto/thing.pb.h
  bazel-genfiles/example/proto/cpp_thing_proto/example/proto/thing.pb.cc
```

This way, we can use `//example/proto:cpp_thing_proto` as a dependency of any
other `cc_library` or `cc_binary` rule as per normal.

> NOTE: the `cpp_proto_library` implicitly calls `cpp_proto_compile`, and we can
access that rule by adding `_pb` at the end of the rule name, like `bazel build
//example/proto:cpp_thing_proto_pb`.


## Migration

For users migrating from the [stackb/rules_proto](https://github.com/stackb/rules_proto)
rules, please see the help at [MIGRATION.md](/docs/MIGRATION.md)


## Developers

### Code Layout

Each language `{lang}` has a top-level subdirectory that contains:

1. `{lang}/README.md`: Generated documentation for the language rules.

1. `{lang}/repositories.bzl`: Macro functions that declare repository rule
   dependencies for that language.

2. `{lang}/{rule}.bzl`: Rule implementations of the form
   `{lang}_{kind}_{type}`, where `kind` is one of `proto|grpc` and `type` is one
   of `compile|library`.

3. `{lang}/BUILD.bazel`: `proto_plugin()` declarations for the available
   plugins for the language.

4. `example/{lang}/{rule}/`: Generated `WORKSPACE` and `BUILD.bazel`
   demonstrating standalone usage of the rules.

5. `{lang}/example/routeguide/`: Example routeguide example implementation, if
   possible.


The repository root directory contains the base rule defintions:

* `plugin.bzl`: A build rule that defines the name, tool binary, and options for
  a particular proto plugin.

* `aspect.bzl`: Contains the implementation of the compilation aspect. This is
  shared by all rules and is the heart of `rules_proto_grpc`; it calls `protoc`
  with a given list of plugins and generates output files.

Additional protoc plugins and their rules are scoped to the github repository
name where the plugin resides.


### Rule Generation

To help maintain consistency of the rule implementations and documentation, all
of the rule implementations are generated by the tool `//tools/rulegen`. Changes
in the main `README.md` should be placed in `tools/rulegen/README.header.md` or
`tools/rulegen/README.footer.md`. Changes to generated rules should be put in
the source files (example: `tools/rulegen/java.go`).


### How-it-works

Briefly, here's how the rules work:

1. Using the `proto_library` graph, an aspect walks through the [`ProtoInfo`](https://docs.bazel.build/versions/master/skylark/lib/ProtoInfo.html)
   providers on the `deps` attribute to `{lang}_{proto|grpc}_compile`. This
   finds all the directly and transitively required proto files, along with
   their options.

2. At each node visited by the aspect, `protoc` is invoked with the relevant
   plugins and options to generate the desired outputs. The aspect uses only
   the generated proto descriptors from the `ProtoInfo` providers.

3. Once the aspect stage is complete, all generated outputs are optionally
   gathered into a final output tree.

4. For `{lang}_{proto|grpc}_library` rules, the generated outputs are then
   aggregated into a language-specific library. e.g a `.so` file for C++.


### Developing Custom Plugins

Generally, follow the pattern seen in the multiple language examples in this
repository.  The basic idea is:

1. Load the plugin rule: `load("@rules_proto_grpc//:plugin.bzl", "proto_plugin")`.
2. Define the rule, giving it a `name`, `options` (not mandatory), `tool`, and
   `outputs`.
3. `tool` is a label that refers to the binary executable for the plugin itself.
4. Choose your output type (pick one!):
    - `outputs`: a list of strings patterns that predicts the pattern of files
      generated by the plugin. For plugins that produce one output file per
      input proto file
    - `out`: the name of a single output file generated by the plugin.
    - `output_directory`: Set to true if your plugin generates files in a
      non-predictable way. e.g. if the output paths depend on the service names.
5. Create a compilation rule and aspect using the following template:

```python
load("@rules_proto_grpc//:plugin.bzl", "ProtoPluginInfo")
load(
    "@rules_proto_grpc//:aspect.bzl",
    "ProtoLibraryAspectNodeInfo",
    "proto_compile_aspect_attrs",
    "proto_compile_aspect_impl",
    "proto_compile_attrs",
    "proto_compile_impl",
)

# Create aspect
example_aspect = aspect(
    implementation = proto_compile_aspect_impl,
    provides = [ProtoLibraryAspectNodeInfo],
    attr_aspects = ["deps"],
    attrs = dict(
        proto_compile_aspect_attrs,
        _plugins = attr.label_list(
            doc = "List of protoc plugins to apply",
            providers = [ProtoPluginInfo],
            default = [
                Label("//<LABEL OF YOUR PLUGIN>"),
            ],
        ),
        _prefix = attr.string(
            doc = "String used to disambiguate aspects when generating outputs",
            default = "example_aspect",
        )
    ),
    toolchains = ["@rules_proto_grpc//protobuf:toolchain_type"],
)

# Create compile rule to apply aspect
_rule = rule(
    implementation = proto_compile_impl,
    attrs = dict(
        proto_compile_attrs,
        deps = attr.label_list(
            mandatory = True,
            providers = [ProtoInfo, ProtoLibraryAspectNodeInfo],
            aspects = [example_compile],
        ),
    ),
)

# Create macro for converting attrs and passing to compile
def example_compile(**kwargs):
    _rule(
        verbose_string = "{}".format(kwargs.get("verbose", 0)),
        merge_directories = True,
        **{k: v for k, v in kwargs.items() if k != "merge_directories"}
    )

```


## License

This project is derived from [stackb/rules_proto](https://github.com/stackb/rules_proto)
under the [Apache 2.0](http://www.apache.org/licenses/LICENSE-2.0) license and
this project therefore maintains the terms of that license. An overview of the
changes can be found at [MIGRATION.md](/docs/MIGRATION.md).


## Contributing

Contributions are very welcome. Please see [CONTRIBUTING.md](/docs/CONTRIBUTING.md)
for further details.
