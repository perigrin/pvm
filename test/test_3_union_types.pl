type NumberOrString = Int|Str;
type OptionalData = HashRef|Undef;
type ComplexUnion = Int|Str|ArrayRef[Int]|HashRef[Str];

my NumberOrString $value = 42;
$value = "hello";

my OptionalData $data = { key => "value" };
$data = undef;
