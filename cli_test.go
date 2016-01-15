package gogrinder

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder"}

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	filename, noExec, noReport, noFrontend, port, err := GetCLI()
	if filename != "loadmodel.json" {
		t.Errorf("Default filename was expected 'loadmodel.json' but was: %s", filename)
	}
	if noExec != false {
		t.Errorf("Default -no-exec was expected false but was: %b", noExec)
	}
	if noReport != false {
		t.Errorf("Default -no-report was expected false but was: %b", noReport)
	}
	if noFrontend != false {
		t.Errorf("Default -no-frontend was expected false but was: %b", noFrontend)
	}
	if port != 3000 {
		t.Errorf("Default port was expected 3000 but was: %d", port)
	}
	if err != nil {
		t.Errorf("Default err was expected nil but was: %s", err)
	}
}

func TestNoExec(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "-no-exec"}

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	_, noExec, _, _, _, err := GetCLI()
	if noExec != true {
		t.Errorf("-no-exec was expected true but was: %b", noExec)
	}
	if err != nil {
		t.Errorf("err was expected nil but was: %s", err)
	}
}

func TestNoFrontend(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "-no-frontend"}

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	_, _, _, noFrontend, _, err := GetCLI()
	if noFrontend != true {
		t.Errorf("-no-frontend was expected true but was: %b", noFrontend)
	}
	if err != nil {
		t.Errorf("err was expected nil but was: %s", err)
	}
}

func TestNoReport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "-no-report"}

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	_, _, noReport, _, _, err := GetCLI()
	if noReport != true {
		t.Errorf("-no-report was expected true but was: %b", noReport)
	}
	if err != nil {
		t.Errorf("err was expected nil but was: %s", err)
	}
}

func TestPort(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "-port", "8888"}

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	_, _, _, _, port, err := GetCLI()
	if port != 8888 {
		t.Errorf("Port was expected 8888 but was: %d", port)
	}
	if err != nil {
		t.Errorf("err was expected nil but was: %s", err)
	}
}

func TestFilename(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", file.Name()}

	filename, _, _, _, _, err := GetCLI()
	if filename != file.Name() {
		t.Errorf("Filename was expected %s but was: %s", file.Name(), filename)
	}
	if err != nil {
		t.Errorf("err was expected nil but was: %s", err)
	}
}

func TestUnknownFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "-unknown"}

	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// prepare the default loadmodel.json file
	f, ferr := os.Create("./loadmodel.json")
	if ferr != nil {
		t.Errorf("problem during default file creation: %s", ferr)
	}
	f.Close()
	defer os.Remove("./loadmodel.json")

	_, _, _, _, _, err := GetCLI()
	if err.Error() != "Command line usage problem." {
		t.Errorf("err was expected %s but was: %s", "Command line usage problem.", err.Error())
	}
}

func TestAdditionalArgument(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", file.Name(), "2nd_argument"}

	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	_, _, _, _, _, err := GetCLI()
	if err.Error() != "Command line usage problem." {
		t.Errorf("err was expected %s but was: %s", "Command line usage problem.", err.Error())
	}
}

func TestFileNotFound(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", "unknown_file.json"}

	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	_, _, _, _, _, err := GetCLI()
	if err.Error() != "File unknown_file.json does not exist." {
		t.Errorf("err was expected %s but was: %s", "File unknown_file.json does not exist.", err.Error())
	}
}
