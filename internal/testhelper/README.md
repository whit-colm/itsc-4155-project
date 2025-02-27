# DO NOT IMPORT THIS PACKAGE

This is a *really* dumb fix I thought after working on this for like 12 hours straight to a problem I shouldn't really be having.

This package provides some helper functions for the various package test suites. **IT SHOULD _NEVER_ BE IMPORTED INTO THE ACTUAL PACKAGE PROPER.**

Also, this should only be used with blackbox tests, otherwise you'll probably end up with cyclic dependencies.