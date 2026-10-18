# The input of repolint() in a consuming repo: the program's source.
filegroup(
    name = "srcs",
    srcs = ["go.mod"] + glob(
        ["*.go"],
        exclude = ["*_test.go"],
    ),
    visibility = ["PUBLIC"],
)

# The input of repolint_test() in a consuming repo: the tests and their data.
filegroup(
    name = "tests",
    srcs = glob(["*_test.go"]),
    visibility = ["PUBLIC"],
)

filegroup(
    name = "testdata",
    srcs = glob(["testdata/**"]),
    visibility = ["PUBLIC"],
)
