package objects

import (
    "time"
    "fmt"
    "strings"
    )


// Represents a git 'tag' object
// The objectHash is the hash of the object being tagged
// The tagType denotes the kind of object being tagged
// The tagName is the name of the tag
// The tagger is the person making tag along with the time stamp (see commit authors)
// The msg is an optional tag message
// The tag object also supports an optional GPG key which is unsupported
type GitTag struct {
    objectHash string
    tagType GitObjectType
    tagName string
    tagger *commitIdentity
    msg string
}
// Returns the size of the tag as a string
func (tag *GitTag) Size() string {
    return fmt.Sprint(tag.computeSize())
}
// Retuns 'tag'
func (tag *GitTag) Type() GitObjectType {
    return Tag
}
// Set the fields of a tag struct using data in src (i.e. parse a tag object)
func (tag *GitTag) Deserialize(src []byte) {
    tag.msg = ""
    lines := strings.Split(string(src), "\n")
    // Loop through all the lines of the tag string
    for _, line := range lines {
        // Break the line into words
        words := strings.Split(line, " ")
        // The first word determines what information the line holds, process accordingly
        switch (words[0]) {
        case "object" : 
            tag.objectHash = words[1]
        case "type" : 
            tag.tagType = GitObjectType(words[1])
        case "tag" : 
            tag.tagName = words[1]
        case "tagger":
            tagger, err := idFromString(strings.Join(words[1:], " "))
            if err != nil {
                panic(err)
            }
            tag.tagger = tagger
        default:
            tag.msg += strings.Join(words, " ") + "\n"
        }
    }
    // Remove the Extra new line that will be added on the tag msg because of how its
    // parsed
    tag.msg = strings.Trim(tag.msg, "\n")
}

// Convert tag struct into a []byte (which is really just a string) ready for writing
// to a file
func (tag *GitTag) Serialize() []byte {
    bytes := make([]byte, 0 , tag.computeSize())
    bytes = append(bytes, "object "...)
    bytes = append(bytes, tag.objectHash...)
    bytes = append(bytes, '\n')

    bytes = append(bytes, "type "...)
    bytes = append(bytes, tag.tagType...)
    bytes = append(bytes, '\n')


    bytes = append(bytes, "tag "...)
    bytes = append(bytes, tag.tagName...)
    bytes = append(bytes, '\n')

    if tag.tagger != nil {
    bytes = append(bytes, "tagger "...)
    bytes = append(bytes, tag.tagger.String()...)
    bytes = append(bytes, '\n')
    }

    bytes = append(bytes, '\n')
    bytes = append(bytes, tag.msg...)
    bytes = append(bytes, '\n')

    return bytes
}

// Returns the string form of a tag
func (tag *GitTag) String() string {
    return string(tag.Serialize())
}
// Compute the overall size of the tag (i.e. the amount of bytes it would take to store
// the tag as a string
func (tag *GitTag) computeSize() int {
    size := 0
    if tag.objectHash != "" {
        //(len("object") = 6) + space + \n = 8
        size += 8 + len(tag.objectHash)
    }

    if tag.tagType != "" {
        //(len("type") = 4) + space + \n = 6
        size += 6 + len(tag.tagType)
    }

    if tag.tagName != ""  {
        //(len("tag") = 3) + space + \n = 5
        size += 5 + len(tag.tagName)
    }


    if tag.tagger != nil { 
        //(len("tagger") = 6) + space + \n = 8
        size += tag.tagger.Size() + 8
    }
    
    // +2 is for the blank \n character before the tag message
    // and the \n inserted after the tag message
    size += len(tag.msg) + 2
    return size
}
// Set the object hash and object type of the tag. 
// Can use struct field assignment, but a function that assigns them together empasizes
// that they represent information about the same object
func (tag *GitTag) SetObject(objType GitObjectType, objHash string) {
    tag.objectHash = objHash
    tag.tagType = objType
}

// Set the tagger (i.e. the person creating the tag)
// the current system time and timezone will be used as a timestamp
func (tag *GitTag) SetTagger(name string, email string) {
    time :=  time.Now()
    _, tz_offset := time.Zone()
    tOffset, err := fromOffset(tz_offset) 
    if err != nil {
        panic (err)
    }
    tagger := commitIdentity{name:name,
                          email:email,
                          time:time.Unix(),
                          timezone:*tOffset}
    tag.tagger = &tagger
}

// Set the tagger and the timezone for the tag
func (tag *GitTag) SetTaggerAndTime(name string,
                                            email string,
                                            timeInUnix int64,
                                            tz_offset int) error {
    tOffset, err := fromOffset(tz_offset)
    if err != nil {
        return err
    }
    tagger := commitIdentity{name:name,
                              email:email,
                              time:timeInUnix,
                              timezone:*tOffset}
    tag.tagger = &tagger
    return nil

}
