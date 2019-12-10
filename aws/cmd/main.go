package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"os/exec"
)

var (
	output string
)

func init() {
	flag.StringVar(&output, "output", "", "The output file of the generated code")
}

func main() {
	flag.Parse()

	if output == "" {
		panic("The 'output' is required")
	}

	f, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = generate(f, functions)
	if err != nil {
		panic(err)
	}
}

func generate(opt io.Writer, fns []Function) error {
	var fnBuff = bytes.Buffer{}

	// Adds the package definition
	err := pkgTmpl.Execute(&fnBuff, nil)
	if err != nil {
		return err
	}

	// Adds the AWSReader interface
	err = awsReaderTmpl.Execute(&fnBuff, fns)
	if err != nil {
		return err
	}

	// Adds the implementation of the functions
	for _, fn := range fns {
		err = fn.Execute(&fnBuff)
		if err != nil {
			return err
		}
	}

	stderr := &bytes.Buffer{}

	// Formats the output using goimports
	cmd := exec.Command("goimports")
	cmd.Stdin = &fnBuff
	cmd.Stdout = opt
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	if serr := stderr.String(); serr != "" {
		return errors.New(serr)
	}

	return nil
}
