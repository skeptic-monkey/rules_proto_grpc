load("//python:python_grpclib_compile.bzl", "python_grpclib_compile")

def python_grpclib_library(**kwargs):
    # Compile protos
    name_pb = kwargs.get("name") + "_pb"
    python_grpclib_compile(
        name = name_pb,
        **{k: v for (k, v) in kwargs.items() if k in ("deps", "verbose")} # Forward args
    )

    # Create python library
    native.py_library(
        name = kwargs.get("name"),
        srcs = [name_pb],
        deps = [
            "@com_google_protobuf//:protobuf_python",
        ] + GRPC_DEPS,
        imports = [name_pb],
        visibility = kwargs.get("visibility"),
    )

GRPC_DEPS = [
    "@rules_proto_grpc_py3_deps//grpclib"
]

# Alias
py_grpclib_library = python_grpclib_library
