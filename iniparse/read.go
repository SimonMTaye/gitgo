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

func ParseIni (file io.Reader) (IniFile, error) {

    fileScanner := bufio.NewScanner(file)
    iniFile := make(IniFile)
    curSection := ""
    
    r := fileScanner.Scan()

    for ; r ; r = fileScanner.Scan() {
        line := fileScanner.Text()
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
