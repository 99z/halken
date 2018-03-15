# GB(C) Test ROMS

These are Blargg's test ROMS from http://gbdev.gg8.se/files/roms/blargg-gb-tests/.

Some are runnable on an emulator that only has CPU + RAM implemented. Test outputs do not even require having graphics of any kind:

```
Text output and the final result are also written to memory at $A000,
allowing testing a very minimal emulator that supports little more than
CPU and RAM. To reliably indicate that the data is from a test and not
random data, $A001-$A003 are written with a signature: $DE,$B0,$61. If
this is present, then the text string and final result status are valid.

$A000 holds the overall status. If the test is still running, it holds
$80, otherwise it holds the final result code.

All text output is appended to a zero-terminated string at $A004. An
emulator could regularly check this string for any additional
characters, and output them, allowing real-time text output, rather than
just printing the final output at the end.
```

This is how I'll be testing halken, especially early on.

The CPU instructions test ROM requires MBC1 to be implemented, but the individual CPU tests do not. I included them for this reason.
