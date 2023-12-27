package problems

import (
	"archive/zip"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
	"github.com/pkg/errors"
	"io"
	"strings"
)

const inKeyword = "in"
const outKeyword = "out"

func getDirsName(r *zip.Reader) (in string, out string, err error) {

	for _, f := range r.File {
		if f.FileInfo().IsDir() {

			splittedName := strings.Split(strings.Trim(f.Name, "/"), "/")
			switch strings.ToLower(splittedName[len(splittedName)-1]) {
			case inKeyword:
				in = f.Name
			case outKeyword:
				out = f.Name
			}
			if in != "" && out != "" {
				break
			}

		}
	}
	if in == "" || out == "" {
		err = pkg.ErrBadRequest
	}
	return
}
func unzip(data io.ReaderAt, size int64) ([]structs.Testcase, error) {

	r, err := zip.NewReader(data, size)
	if err != nil {
		return nil, err
	}

	in, out, err := getDirsName(r)
	if err != nil {
		return nil, err
	}

	testCases := make(map[string]structs.Testcase)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		reader, err := f.OpenRaw()
		if err != nil {
			return nil, errors.WithStack(errors.WithMessage(err, "error on open raw file"))
		}
		dataRaw, err := io.ReadAll(reader)
		data := strings.TrimSpace(string(dataRaw))
		if err != nil {
			return nil, errors.WithStack(errors.WithMessage(err, "error on read file"))
		}

		if strings.HasPrefix(f.Name, in) {
			testName := strings.Replace(f.Name, inKeyword, "", -1)
			t := testCases[testName]
			t.Input = data
			testCases[testName] = t
		}
		if strings.HasPrefix(f.Name, out) {
			testName := strings.Replace(f.Name, outKeyword, "", -1)
			t := testCases[testName]
			t.ExpectedOutput = data
			testCases[testName] = t
		}
	}

	ans := make([]structs.Testcase, len(testCases))
	i := 0
	for _, v := range testCases {
		ans[i] = v
		i++
	}
	return ans, nil
}
