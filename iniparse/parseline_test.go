package iniparse

import "testing"

func TestParseLineSection(t *testing.T) {
    line := "    [Header]"
    sec, blank, linetype := parseLine(line)
    if sec != "Header" || blank != "" || linetype != SectionLine {
        t.Errorf(
            "Error parsing section line.\nExpected: 'Header', %d\nGot: %s, %d", 
            SectionLine, sec, linetype)
    }

    // Parse Line doesn't test for unclosed braces, refer to ini spec on Wikipedia
    line = "[_lowercase"
    sec, blank, linetype = parseLine(line)
    if sec != "_lowercase"  || linetype != SectionLine {
        t.Errorf(
            "Error parsing section line.\nExpected: '_lowercase', %d\nGot: %s, %d", 
            SectionLine, sec, linetype)
    }
}


func TestParseLineComment(t *testing.T) {
    line := "#Some comment"
    com, _,linetype := parseLine(line)
    if com != line || linetype != CommentLine {
        t.Errorf(
            "Error parsing comment line.\nExpected:\n\t%s\n\tLine Type: %d\n" +
            "Got:\n\t%s\n\tLine Type: %d", 
            line, CommentLine, com, linetype)
    }
    line = ";Some comment"
    com, _,linetype = parseLine(line)
    if com != line || linetype != CommentLine {
        t.Errorf(
            "Error parsing comment line.\nExpected:\n\t%s\n\tLine Type: %d\n" +
            "Got:\n\t%s\n\tLine Type: %d", 
            line, CommentLine, com, linetype)
    }
}

func TestParseLineEntry(t *testing.T) {
    line := "bestcolor=purple"
    name,val,linetype := parseLine(line)

    if name != "bestcolor" || val != "purple" || linetype != EntryLine {
        t.Errorf(
            "Error parsing entry line.\nExpected:\n\tName: %s\n\tValue: %s\n\tLine Type: %d\n" +
            "Got:\n\tName: %s\n\tValue: %s\n\tLine Type: %d" ,
            "bestcolor","purple", EntryLine, name, val, linetype)
        
    }
}
