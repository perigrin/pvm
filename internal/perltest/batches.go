// ABOUTME: The batches of core modules held to perl, named once and shared by
// ABOUTME: internal/parse (prototypes, exports) and internal/infer (return types).

package perltest

// XSUtilities are the XS utility modules declared as one batch.
var XSUtilities = []string{"Scalar::Util", "List::Util", "Cwd", "Data::Dumper", "MIME::Base64", "Storable"}

// PureExporters is the pure-Perl exporters batch (01a1464e-db4c): every
// module whose .tpm declares the names a `use` of it installs. Symbol
// was held back (see docs/rfcs/0001-type-libraries.md and issue
// 01a156e4): its qualify, declared accurately, made IO/Handle.pm's `tell
// qualify($_[0], caller) if ref($_[0])` a new TestPSCCheckLibBaseline
// diagnostic perl -w did not confirm, since $_[0] there is always a ref.
// Narrowing $_[0] after that guard and declaring qualify a multi by
// argument type (Ref keeps that Ref, Str gives a Str) fixed the
// narrowing half; the other half was a types.Select gap (select.go,
// SelectShaped's coercion retry and its Unknown bypass), fixed in
// 01a156e4 too, caller's own List not fitting qualify's $pkg otherwise
// discarding $_[0]'s proven Ref along with it.
var PureExporters = []string{
	"File::Basename", "File::Find", "File::Path", "Errno", "Symbol",
	"Text::ParseWords", "Class::Struct", "File::Copy", "Time::Local", "Getopt::Std",
}

// XSLibraries is the XS-library batch (01a1464e-dbd3): Encode,
// Time::HiRes, Compress::Raw::Zlib and SDBM_File.
var XSLibraries = []string{"Encode", "Time::HiRes", "Compress::Raw::Zlib", "SDBM_File"}

// OOModules is the OO-only batch (01a1464e-dd27): each declares methods,
// not exports, since measured on 5.42.0 a bare `use` of every one but
// IO::Socket::IP and Locale::Maketext::Simple installs nothing.
var OOModules = []string{
	"File::Spec", "File::Spec::Unix", "IO::Socket::IP", "IO::Select",
	"Term::ReadLine", "Tie::Hash", "HTTP::Tiny", "Locale::Maketext::Simple",
}

// VariableExporters is the variable-exporters batch (01a1464e-dca3):
// every module below exports a plain Perl variable, measured on
// 2026-10-16.
var VariableExporters = []string{
	"Config", "Getopt::Long", "Test::More", "Text::Wrap", "ExtUtils::MakeMaker", "File::stat",
}

// IOExporters is the IO and file-handling batch (01a1464e-db76):
// FileHandle, IO::Handle, IO::File, IO::Socket and the five plain
// exporters File::Spec::Functions, File::Temp, Archive::Tar,
// ExtUtils::Manifest and Net::Ping.
var IOExporters = []string{
	"FileHandle", "IO::Handle", "IO::File", "IO::Socket",
	"File::Spec::Functions", "File::Temp", "Archive::Tar", "ExtUtils::Manifest", "Net::Ping",
}

// Standalone are the core modules each measured alone rather than as part
// of a batch above: Carp and POSIX (each large enough on its own), the
// constant exporters Fcntl and Socket (measured together but under no
// shared batch name), and Exporter, XSLoader and DynaLoader -- the three
// infrastructure modules RFC 0001 permits to ship unshipped (see
// docs/contributing/core-module-declarations.md, "Modules other modules
// depend on"), each measured and held separately so its own criteria
// command names it.
var Standalone = []string{"Carp", "POSIX", "Fcntl", "Socket", "Exporter", "XSLoader", "DynaLoader"}

// AllHeld is every module name in every batch above: the complete set the
// top-50 core-module declarations are held to perl by, across both
// internal/parse and internal/infer. TestHoldListsCoverEveryDeclaration
// (internal/parse) checks it against the ranking that chooses the 50, so
// a module the ranking adds or drops is caught here rather than only in
// whichever package's batch a worker remembered to update.
func AllHeld() []string {
	var out []string
	for _, batch := range [][]string{
		XSUtilities, PureExporters, XSLibraries, OOModules, VariableExporters, IOExporters, Standalone,
	} {
		out = append(out, batch...)
	}
	return out
}
