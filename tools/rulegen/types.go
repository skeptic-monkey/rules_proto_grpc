package main

import (
	"text/template"
)


// Language represents one directory in this repo
type Language struct {
	// Directory in the repo where this language is rooted.  Typically this is
	// the same as the name
	Dir string

	// Name of the language
	Name string

	// The display name of the language
	DisplayName string

	// Workspace usage
	WorkspaceExample string

	// List of rules
	Rules []*Rule

	// Additional nodes about the language
	Notes *template.Template

	// Bazel build flags required / suggested
	Flags []*Flag

	// Additional CI-specific env vars in the form "K=V"
	PresubmitEnvVars map[string]string

	// Flag indicating if the merge_directories flag should be set to false for the generated rule
	SkipDirectoriesMerge bool

	// Platforms for which to skip testing this lang
	// The special value 'all' will skip app platforms
	SkipTestPlatforms []string
}


type Rule struct {
	// Name of the rule
	Name string

	// Base name of the rule (typically the lang name)
	Base string

	// Kind of the rule (proto|grpc)
	Kind string

	// Description
	Doc string

	// Temmplate for workspace
	WorkspaceExample *template.Template

	// Template for build file
	BuildExample *template.Template

	// Template for bzl file
	Implementation *template.Template

	// List of attributes
	Attrs []*Attr

	// List of plugins
	Plugins []string

	// Not expected to be functional
	Experimental bool

	// Bazel build flags required / suggested
	Flags []*Flag

	// Additional CI-specific env vars in the form "K=V"
	PresubmitEnvVars map[string]string

	// Platforms for which to skip testing this rule, overrides language level
	// The special value 'all' will skip app platforms
	SkipTestPlatforms []string
}


// Flag captures information about a bazel build flag.
type Flag struct {
	Category    string
	Name        string
	Value       string
	Description string
}


type Attr struct {
	Name      string
	Type      string
	Default   string
	Doc       string
	Mandatory bool
}


type Plugin struct {
	Tool    string
	Options []string
}


type ruleData struct {
	Lang *Language
	Rule *Rule
}
