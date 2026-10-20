# ABOUTME: Call cases for Time::HiRes's subs, which the return-type oracle calls to hold Time::HiRes.tpm to perl.
# ABOUTME: Each sub maps to a list of argument lists, one call per list; every path stays relative.
#
# The driver only requires Time::HiRes, never imports it, and measured,
# a bare require leaves every one of these generated on demand only --
# `exists &Time::HiRes::CLOCK_REALTIME` is false until something asks
# for it by name. This `use` is that ask, priming every name below so
# the driver's own by-name symbol lookup finds it.
use Time::HiRes qw(
    CLOCKS_PER_SEC CLOCK_BOOTTIME CLOCK_HIGHRES CLOCK_MONOTONIC
    CLOCK_MONOTONIC_COARSE CLOCK_MONOTONIC_FAST CLOCK_MONOTONIC_PRECISE
    CLOCK_MONOTONIC_RAW CLOCK_PROCESS_CPUTIME_ID CLOCK_PROF CLOCK_REALTIME
    CLOCK_REALTIME_COARSE CLOCK_REALTIME_FAST CLOCK_REALTIME_PRECISE
    CLOCK_REALTIME_RAW CLOCK_SECOND CLOCK_SOFTTIME CLOCK_THREAD_CPUTIME_ID
    CLOCK_TIMEOFDAY CLOCK_UPTIME CLOCK_UPTIME_COARSE CLOCK_UPTIME_FAST
    CLOCK_UPTIME_PRECISE CLOCK_UPTIME_RAW CLOCK_VIRTUAL
    ITIMER_PROF ITIMER_REAL ITIMER_REALPROF ITIMER_VIRTUAL TIMER_ABSTIME
    d_clock d_clock_getres d_clock_gettime d_clock_nanosleep d_futimens
    d_getitimer d_gettimeofday d_hires_stat d_hires_utime d_nanosleep
    d_setitimer d_ualarm d_usleep d_utimensat
    alarm clock_getres clock_gettime gettimeofday
    getitimer setitimer lstat sleep stat time tv_interval
    usleep
);
# clock, clock_nanosleep, nanosleep, ualarm and utime each die
# "unimplemented in this platform" the moment Windows's perl primes them
# by name (measured), which would otherwise crash this file's own load
# before any other name above got primed; requesting each on its own,
# guarded, keeps every other name's priming intact. Windows's own oracle
# run then fails calling whichever of these it is actually asked to call,
# for the same reason, rather than this file failing to load at all.
for my $name (qw(clock clock_nanosleep nanosleep ualarm utime)) {
    eval { Time::HiRes->import($name) };
}
open my $fh, '>', 'hr_file' or die $!;
close $fh;
(
    CLOCKS_PER_SEC        => [ [] ],
    CLOCK_BOOTTIME        => [ [] ],
    CLOCK_HIGHRES         => [ [] ],
    CLOCK_MONOTONIC       => [ [] ],
    CLOCK_MONOTONIC_COARSE  => [ [] ],
    CLOCK_MONOTONIC_FAST    => [ [] ],
    CLOCK_MONOTONIC_PRECISE => [ [] ],
    CLOCK_MONOTONIC_RAW     => [ [] ],
    CLOCK_PROCESS_CPUTIME_ID => [ [] ],
    CLOCK_PROF              => [ [] ],
    CLOCK_REALTIME          => [ [] ],
    CLOCK_REALTIME_COARSE   => [ [] ],
    CLOCK_REALTIME_FAST     => [ [] ],
    CLOCK_REALTIME_PRECISE  => [ [] ],
    CLOCK_REALTIME_RAW      => [ [] ],
    CLOCK_SECOND            => [ [] ],
    CLOCK_SOFTTIME          => [ [] ],
    CLOCK_THREAD_CPUTIME_ID => [ [] ],
    CLOCK_TIMEOFDAY         => [ [] ],
    CLOCK_UPTIME            => [ [] ],
    CLOCK_UPTIME_COARSE     => [ [] ],
    CLOCK_UPTIME_FAST       => [ [] ],
    CLOCK_UPTIME_PRECISE    => [ [] ],
    CLOCK_UPTIME_RAW        => [ [] ],
    CLOCK_VIRTUAL           => [ [] ],
    ITIMER_PROF      => [ [] ],
    ITIMER_REAL      => [ [] ],
    ITIMER_REALPROF  => [ [] ],
    ITIMER_VIRTUAL   => [ [] ],
    TIMER_ABSTIME    => [ [] ],

    d_clock           => [ [] ],
    d_clock_getres    => [ [] ],
    d_clock_gettime   => [ [] ],
    d_clock_nanosleep => [ [] ],
    d_futimens        => [ [] ],
    d_getitimer       => [ [] ],
    d_gettimeofday    => [ [] ],
    d_hires_stat      => [ [] ],
    d_hires_utime     => [ [] ],
    d_nanosleep       => [ [] ],
    d_setitimer       => [ [] ],
    d_ualarm          => [ [] ],
    d_usleep          => [ [] ],
    d_utimensat       => [ [] ],

    # 0 cancels any pending alarm/itimer rather than scheduling one, and
    # the sleep/nanosleep/usleep cases sleep zero seconds (a thread
    # yield, per perldoc Time::HiRes), so none of these blocks or sends
    # a real signal.
    alarm           => [ [0] ],
    clock           => [ [] ],
    clock_getres    => [ [] ],
    clock_gettime   => [ [] ],
    clock_nanosleep => [ [ 0, 0 ] ],
    gettimeofday    => [ [] ],
    getitimer       => [ [0] ],
    setitimer       => [ [ 0, 0 ] ],
    lstat           => [ ['hr_file'] ],
    nanosleep       => [ [0] ],
    sleep           => [ [0] ],
    stat            => [ ['hr_file'] ],
    time            => [ [] ],
    tv_interval     => [ [ [ 0, 0 ], [ 1, 500_000 ] ] ],
    ualarm          => [ [0] ],
    usleep          => [ [0] ],
    utime           => [ [ time(), time(), 'hr_file' ] ],
);
