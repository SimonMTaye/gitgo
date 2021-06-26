package objects

import ( 
    "fmt"
    "strings"
    "time"
    "strconv"
    "errors"
)
//Commit object format:
//Header
//tree [tree-hash]\n
//parent [commit-hash]\n (optional)
//author [Name] [email] [time-stamp] [time-zone] (person who wrote the code)
//committer (same format as author)\n\n (person who is committing the code)
//PGP signature (not implemented for now)
//Commit msg + \n

// time-zone format: [+/-][0000]
// where 0000 indicates offset from UTC; +/- indicates postive/negative difference
// timestamp: number of seconds since Jan 1, 1970 00:00

// A commit object
type GitCommit struct {
    treeHash string
    parentHash string
    author *commitIdentity
    committer *commitIdentity
    msg string
}
// Struct that indentifies a committer or author of a commit and when that commit was made
type commitIdentity struct {
    name string
    email string
    time int64
    timezone timeOffset
}
// Data for storing/processing time stamps stored in commits
type timeOffset struct {
    postive bool
    hours int
    mins int
}
// Return a timeOffset sturct from an offset in seconds
// Does not validate input (i.e. hours could be greater than 12)
func fromOffset(offset int) (*timeOffset, error) {
    postive := true
    if offset < 0 {
        postive = false
        offset = -offset
    }
    hours := offset / 3600
    if hours > 12 {
        if postive {
            return nil, 
            errors.New(
                fmt.Sprintf(
                    "The offset %d corresponds to a larger than 12 hour time difference",
                    offset))
        } else {
            return nil, 
            errors.New(
                fmt.Sprintf(
                    "The offset %d corresponds to a larger than 12 hour time difference",
                    -offset))

        }
    }
    mins := offset % 3600
    mins = mins / 60
    return &timeOffset{hours:hours, mins:mins, postive:postive}, nil
} 
// Return a timeOffset struct based on a timezone string as stored by git
func fromOffsetString(offsetString string) (*timeOffset, error) {
    if len(offsetString) != 5 {
        return nil, errors.New("Offset size string is not as expected (5): " + offsetString)
    }
    postive := true
    switch (offsetString[0]) {
    case '+' : postive = true
    case '-' : postive = false
    default  : return nil, errors.New("Expected a +/- as the first char of the string: " + offsetString)
    }
    // two chars for the hour, it shouldn't be larger
    hours, err := strconv.Atoi(offsetString[1:3])
    if err != nil {
        return nil, err
    }
    // two chars for the mins, it shouldn't be larger
    mins, err := strconv.Atoi(offsetString[3:])
    if err != nil {
        return nil, err
    }
    return &timeOffset{hours:hours, mins:mins, postive:postive}, nil
}

func (to *timeOffset) Stringer() string {
    offsetString := make([]byte, 0, 5)
    if to.postive {
        offsetString = append(offsetString, '+')
    } else {
        offsetString = append(offsetString, '-')
    }
    hours := fmt.Sprint(to.hours)

    if len(hours) > 1 {
        offsetString = append(offsetString, hours[0:]...)
    } else {
        offsetString = append(offsetString, '0')
        offsetString = append(offsetString, hours...)
    }

    mins := fmt.Sprint(to.mins)
    if len(mins) > 1 {
        offsetString = append(offsetString, mins[0:]...)
    } else {
        offsetString = append(offsetString, '0')
        offsetString = append(offsetString, mins...)
    }
    
    return string(offsetString)
}

// Returns a commitIdentity struct from a well formed string
// A well formed string is one in the form:
// [name] <[email]> [timestamp] [timezone]
// The meaning of [timestamp] and [timezone] is found above this function
func idFromString (idString string) (*commitIdentity, error) {
    words := strings.Split(idString, " ")
    // Find the email word, everything before is the name
    emailPos := 0
    for i, word := range words {
        if strings.HasPrefix(word, "<") && strings.HasSuffix(word, ">") {
            emailPos = i
            break
        }
    }
    // emailPos will only be 0 if a word wrapped in '< >' is not found (i.e. the string
    // has no email in the format stored by git
    if emailPos == 0 {
        return nil, errors.New(fmt.Sprintf("String is badly formed: %s", idString))
    }
    name := strings.Join(words[0:emailPos], " ")
    //Remove the < and > used to indentify the email
    email := strings.Trim(words[emailPos], "<>")
    timeUnix, err := strconv.Atoi(words[emailPos + 1])
    if err != nil {
        return nil, err
    }

    timeOffset, err := fromOffsetString(words[emailPos + 2])
    if err != nil {
        return nil, err
    }
    return &commitIdentity{name:name, email:email, time:int64(timeUnix), timezone: *timeOffset}, nil
}
// Returns a string form of an author/committer
func (id *commitIdentity) Stringer() string {
    return id.name + " <" + id.email + "> " + fmt.Sprint(id.time) + " " + id.timezone.Stringer()
}

// Returns the size of the commitIdentity in string format
// implementation is a bit wasteful
// optimize
func (id *commitIdentity) Size() int {
    return len(id.Stringer())
}

// Return 'commit'
func (commit *GitCommit) Type() GitObjectType {
    return Commit
}

// Return size of commit data in bytes
func (commit *GitCommit) Size() string {
    return fmt.Sprint(commit.computeSize())
}
// Process a commit string (stored in a commit file) and sets and object field based on
// the data
func (commit *GitCommit) Deserialize(src []byte) {
    commit.msg = ""
    lines := strings.Split(string(src), "\n")
    // Loop through all the lines of the commit string
    for _, line := range lines {
        // Break the line into words
        words := strings.Split(line, " ")
        // The first word determines what information the line holds, process accordingly
        switch (words[0]) {
        case "tree" : 
            commit.treeHash = words[1]
        case "parent":
            commit.parentHash = words[1]
        case "author": 
            author, err := idFromString(strings.Join(words[1:], " "))
            if err != nil {
                panic(err)
            }
            commit.author = author
        case "committer":
            committer, err := idFromString(strings.Join(words[1:], " "))
            if err != nil {
                panic(err)
            }
            commit.committer = committer
        default:
            commit.msg += strings.Join(words, " ") + "\n"
        }
    }
    // Remove the Extra new line that will be added on the commit msg because of how its
    // parsed
    commit.msg = strings.Trim(commit.msg, "\n")
}

// Convert commit struct into a []byte (which is really just a string) ready for writing
// to a file
func (commit *GitCommit) Serialize() []byte {
    bytes := make([]byte, 0 , commit.computeSize())
    bytes = append(bytes, "tree "...)
    bytes = append(bytes, commit.treeHash...)
    bytes = append(bytes, '\n')

    if commit.author != nil {
    bytes = append(bytes, "author "...)
    bytes = append(bytes, commit.author.Stringer()...)
    bytes = append(bytes, '\n')
    }

    if commit.committer != nil {
    bytes = append(bytes, "committer "...)
    bytes = append(bytes, commit.author.Stringer()...)
    bytes = append(bytes, '\n')
    }

    bytes = append(bytes, '\n')
    bytes = append(bytes, commit.msg...)
    // TODO It looks like a new line is added to the final git commit object
    // Check if this is because of pretty pretty printing (unlikely because the size
    // also reflects this)
    bytes = append(bytes, '\n')

    return bytes
}
// Calculates the size of a commit object in bytes, updates the size field in the struct
// and returns the size
func (commit *GitCommit) computeSize() int {
    size := 0
    if commit.treeHash != "" {
        //(len("tree") = 4) + space + \n = 6
        size += 6 + len(commit.treeHash)
    }

    if commit.parentHash != "" {
        //(len("parent") = 6) + space + \n =  8
        size += 8 + len(commit.parentHash)
    }

    if commit.author != nil { 
        //(len("author") = 6) + space + \n = 8
        size += commit.author.Size() + 8
    }
    if commit.committer != nil {
        //(len("committer") = 9) + space + \n = 11
        size += commit.committer.Size() + 11
    }
    // +2 is for the blank \n character before the commit message
    // and the \n inserted after the commit message
    size += len(commit.msg) + 2
    return size

}
// Set the commit author (i.e. the original author of the code in the commit)
// the system time and timezone will be used for those fields
func (commit *GitCommit) SetAuthor(name string, email string) {
    time :=  time.Now()
    _, tz_offset := time.Zone()
    tOffset, err := fromOffset(tz_offset) 
    if err != nil {
        panic (err)
    }
    author := commitIdentity{name:name,
                          email:email,
                          time:time.Unix(),
                          timezone:*tOffset}
    commit.author = &author
}

// Set the commit committer (i.e. the person committing the code)
// the system time and timezone will be used for those fields
func (commit *GitCommit) SetCommitter(name string, email string) {
    time := time.Now()
    _, tz_offset := time.Zone()
    tOffset, err := fromOffset(tz_offset) 
    if err != nil {
        panic (err)
    }
    committer := commitIdentity{name:name,
                          email:email,
                          time:time.Unix(),
                          timezone:*tOffset}
    commit.committer = &committer
}

// Set the commit author (i.e. the original author of the code in the commit)
// tz_offset denotes the offset of time in seconds from UTC
func (commit *GitCommit) SetAuthorAndTime(name string,
                                          email string,
                                          timeInUnix int64,
                                          tz_offset int) error {
    tOffset, err := fromOffset(tz_offset)
    if err != nil {
        return err
    }
    author := commitIdentity{name:name,
                          email:email,
                          time:timeInUnix,
                          timezone:*tOffset}
    commit.author = &author
    return nil
}

// Set the commit committer (i.e. the person committing the code)
// tz_offset denotes the offset of time in seconds from UTC
func (commit *GitCommit) SetCommitterAndTime(name string,
                                            email string,
                                            timeInUnix int64,
                                            tz_offset int) error {
    tOffset, err := fromOffset(tz_offset)
    if err != nil {
        return err
    }
    committer := commitIdentity{name:name,
                              email:email,
                              time:timeInUnix,
                              timezone:*tOffset}
    commit.committer = &committer
    return nil

}
