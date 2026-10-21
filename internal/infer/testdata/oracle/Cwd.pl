# ABOUTME: Call cases for Cwd's subs, which the return-type oracle calls to hold Cwd.tpm to perl.
# ABOUTME: every call below runs from a directory removed out from under the process.

# getcwd and fastcwd are undef when the current directory no longer
# exists (perldoc Cwd: "On error returns C<undef>, with C<$!> set"),
# measured by removing the directory the oracle calls run in: mkdir and
# chdir into "doomed", then rmdir it by its still-resolvable relative
# name from the parent, leaving the process sitting in a directory with
# no name any more. Measured on 5.42.0, every other sub here still
# answers without dying from that state: cwd and fastgetcwd give the
# empty string, chdir(".") is 1 (it is already there), and abs_path,
# fast_abs_path, realpath and fast_realpath are each undef for "." --
# every one of those already within its own declared type.
mkdir 'doomed' or die "mkdir: $!";
chdir 'doomed' or die "chdir: $!";
# Windows refuses to remove a directory a process sits in (CI job on
# 66180e7e: "rmdir: Permission denied"), so there the calls run from a
# directory that still exists, every answer a Str within its declared
# type; the undef answers are measured on Linux and macOS.
rmdir '../doomed' or $^O eq 'MSWin32' or die "rmdir: $!";

(
    cwd           => [ [] ],
    getcwd        => [ [] ],
    fastcwd       => [ [] ],
    fastgetcwd    => [ [] ],
    chdir         => [ ['.'] ],
    abs_path      => [ ['.'] ],
    fast_abs_path => [ ['.'] ],
    realpath      => [ ['.'] ],
    fast_realpath => [ ['.'] ],
);
