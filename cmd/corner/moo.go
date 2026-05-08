package main

// mooASCII is a small easter-egg shown when the user passes -M/--moo. It is
// intentionally local to the CLI binary instead of living in internal/config,
// where it has nothing to do with configuration concerns.
const mooASCII = `
                (__)
                (oo)
           /-----\/
          / |   ||
        *  /\--/\
           ~~  ~~
`
