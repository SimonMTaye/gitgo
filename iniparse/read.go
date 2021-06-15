package iniparse

import (
    "io"
    "strings"
    "bufio"
)

// The type of line parseLine can return
type LineType int
const (
    CommentLine LineType = 0
    SectionLine LineType = 1
    EntryLine LineType = 2
    BadLine LineType = 3
)

// Error type for when the ini file parsing fails
type ErrBadLine struct {
    line string
}

func (e ErrBadLine) Error() string {
    slice := make([]string, 2)
    slice[0] = "Error parsing line"
    slice[1] = e.line
    return strings.Join(slice, "\n")
}

//Takes an io.Reader and parses into an iniFile
//If an error is encoutered, returns the parsed iniFile so far and the error
//Otherwise, returns the iniFile and nil
func ParseIni (file io.Reader) (IniFile, error) {

    fileScanner := bufio.NewScanner(file)
    iniFile := make(IniFile)
    curSection := ""
    
    ok := fileScanner.Scan()

    for ; ok ; ok = fileScanner.Scan() {
        line := fileScanner.Text()
        //val: value, opt: optional (is empty unless parseLine finds a entry line
        val, opt, lineType := parseLine(line)

        if lineType == SectionLine {
            curSection = val
        } else if lineType == EntryLine {

            if iniFile[curSection] ==  nil {
                iniFile[curSection] = make(Section)
            }

            iniFile[curSection][val] = opt
        } else if lineType == BadLine {
            err := ErrBadLine{line: line}
            return iniFile, err
        }
    }

    return iniFile, fileScanner.Err()
}


//Parses a line of an iniFile and returns the line type as well as two strings
//For a 'EntryLine', the two strings correspond to the key-value pair, respectively.
//For 'CommentLine' and 'SectionLine', the second string is "" with the the first string
//containing the entire line for comments and the section name between the '[' ']' for
//section line.
//Space at the end of strings is trimmed before being returned
func parseLine (line string) (string, string, LineType) {
    line = strings.Trim(line, " ")

    if strings.HasPrefix(line, "[") {
        header := strings.Trim(line, "[]")
        header = strings.Trim(header, " ")
        return header ,"",SectionLine
    }

    //# is used conventionally as a comment delimeter but not always
    if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
        return line, "", CommentLine
    }

    val := strings.Split(line, "=")
    if len(val) > 1 {
        return strings.Trim(val[0], " "), strings.Trim(val[1], " ") , EntryLine
    }

    return line, "", BadLine    
}
