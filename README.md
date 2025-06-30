## Building

To build a redistributable, production mode package, use `wails build -tags=build,production,webkit2_41`.


## TO INSTALL
To install user `go install -tags=build,production,webkit2_41`

to test application you must
1. build the application
2. install the application
3. use a command that starts with "ez-utils" (e.g. `ez-utils edit population <path to file>`)




first run wails build -tags=build,production,webkit2_41
then run go install -tags=build,production,webkit2_41
then test



wails build -tags="build,production,webkit2_41" -nopackage -clean -dryrun