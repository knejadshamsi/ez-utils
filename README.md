## Building

To build a redistributable, production mode package, use `wails build -tags=wails,production,webkit2_41`.


## TO INSTALL
To install user `go install -tags=wails,production,webkit2_41`

to test application you must
1. build the application
2. install the application
3. use a command that starts with "ez-utils" (e.g. `ez-utils edit population <path to file>`)




first run wails build -tags=wails,production,webkit2_41
then run go install -tags=wails,production,webkit2_41
then test: cd test && ez-utils edit popultation test_pop.xml


# Clean build to ensure fresh bindings
  wails build -clean -tags=wails,production,webkit2_41

  # Install
  go install -tags=wails,production,webkit2_41

  # Test CLI commands (should NOT launch GUI)
  ez-utils scale
  ez-utils create
  
  # Test Transit Vehicle TUI (creates/edits transit vehicles)
  ez-utils create tv transitVehicles.xml

  # Test GUI command (should launch GUI)
  ez-utils edit population test_pop.xml

  # run in dev mode:
  wails dev -tags='wails,!production,webkit2_41'