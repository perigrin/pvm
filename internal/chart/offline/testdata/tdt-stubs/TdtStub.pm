# ABOUTME: Stub module the -DT capture sample loads; its tokens must not reach the sample's capture.
# ABOUTME: Declares a sentinel sub the sample never spells, so a leaked module token is detectable.
package TdtStub;
use strict;
sub TDT_SENTINEL_word { BEGIN { eval "1" } 1 }
1;
