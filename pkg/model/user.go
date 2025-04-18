package model

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"

	"github.com/google/uuid"
)

const UserApiVersion string = "user.itsc-4155-group-project.edu.whits.io/v1alpha1"

type User struct {
	ID          uuid.UUID `json:"id"`
	GithubID    string    `json:"github_id"`
	DisplayName string    `json:"name"`
	Pronouns    string    `json:"pronouns"`
	Username    Username  `json:"username"`
	Email       string    `json:"email"`
	Avatar      uuid.UUID `json:"bref_avatar"`
	Admin       bool      `json:"admin"`
}

func (u User) APIVersion() string {
	return UserApiVersion
}

func (u User) ToAuthor() CommentUser {
	return CommentUser{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Pronouns:    u.Pronouns,
		Username:    u.Username,
		Avatar:      u.Avatar,
	}
}

var (
	// Validate the general structure of a username. This does not
	// actually validate a a username, but should be used as a
	// preliminary check that a string *could* be a username at
	// its highest level.
	UsernameComponentsRe *regexp.Regexp

	// Validate the general structure of a handle.
	//
	// generalHandleRe is responsible for length and globally banned
	// character sets.
	//
	// A handle is 2-32 characters, contains any characters except
	// #, @, or newline, and does not have a leading or trailing space.
	// Due to how handles have to be managed, handle regular
	// expressions expect to be given handles as the only thing in
	// their sequence to validate
	//
	// A valid handle matches [generalHandleRe] and does not match
	// either [lHandleWhtespRe] or [rHandleWhtespRe]
	generalHandleRe *regexp.Regexp
	// Validate potential leading whitespace of a handle.
	//
	// rHandleWhtespRe is responsible for trailing whitespace.
	//
	// A handle is 2-32 characters, contains any characters except
	// #, @, or newline, and does not have a leading or trailing space.
	// Due to how handles have to be managed, handle regular
	// expressions expect to be given handles as the only thing in
	// their sequence to validate
	//
	// A valid handle matches [generalHandleRe] and does not match
	// either [lHandleWhtespRe] or [rHandleWhtespRe]
	lHandleWhtespRe *regexp.Regexp
	// Validate potential trailing whitespace of a handle.
	//
	// rHandleWhtespRe is responsible for trailing whitespace
	//
	// A handle is 2-32 characters, contains any characters except
	// #, @, or newline, and does not have a leading or trailing space.
	// Due to how handles have to be managed, handle regular
	// expressions expect to be given handles as the only thing in
	// their sequence to validate
	//
	// A valid handle matches [generalHandleRe] and does not match
	// either [lHandleWhtespRe] or [rHandleWhtespRe]
	rHandleWhtespRe *regexp.Regexp

	ReservedHandles []string = []string{"system", "deleted", "invalid"}
)

func init() {
	// Capture group 1 is the handle, 2 is the discrim
	UsernameComponentsRe = regexp.MustCompile(`^([^@#\n]{2,32})#(\d{4})$`)

	generalHandleRe = regexp.MustCompile(`^[^@#\n]{2,32}$`)
	lHandleWhtespRe = regexp.MustCompile(`^\s+.*$`)
	rHandleWhtespRe = regexp.MustCompile(`^.*\s+$`)
}

// A username is a human-readable unique identifier for a user and
// comprises two parts: a handle (some text) and a discriminator (4
// digits), separated by a `#` symbol. There are a handful of specific
// constraints on this structure, however:
//
//  1. The handle is 2-32 unicode characters, except for control codes,
//     newlines, or the characters `#` or `@`.
//  2. The handle neither starts nor ends with whitespace
//  3. The handle is not a reserved handle. Currently the reserved
//     handles are `system` and `deleted user`
//  4. The discriminator is not `0000`; this should not be blocked by
//     schema, but is instead reserved for the aforementioned reserved
//     handles.
type Username struct {
	handle        string
	discriminator int16
}

// Create a Username by coercing a string.
//
// This returns an error if the string cannot be parsed or
func UsernameFromString(un string) (Username, error) {
	components := UsernameComponentsRe.FindStringSubmatch(un)

	if len(components) != 3 {
		return Username{},
			fmt.Errorf(
				"incorrect username components count: want `3`, have `%d`",
				len(components),
			)
	}
	handle := components[1]
	if !generalHandleRe.MatchString(handle) {
		return Username{}, fmt.Errorf(
			"invalid handle, must not contain `#`, `@`, or newline",
		)
	} else if lHandleWhtespRe.MatchString(handle) {
		return Username{}, fmt.Errorf(
			"invalid handle, must not have leading whitespace",
		)
	} else if rHandleWhtespRe.MatchString(handle) {
		return Username{}, fmt.Errorf(
			"invalid handle, must not have trailing whitespace",
		)
	}

	discriminator, err := strconv.Atoi(components[2])
	if err != nil {
		return Username{}, fmt.Errorf("cast discriminator: %w", err)
	}

	if uname, err := UsernameFromComponents(
		handle, discriminator,
	); err != nil {
		return Username{}, err
	} else {
		return uname, nil
	}
}

func UsernameFromComponents[I int16 | int](handle string, discriminator I) (Username, error) {
	// The actual discriminator, we have to suss out generics first
	var d int16
	switch v := any(discriminator).(type) {
	case string, int:
		var dI int
		switch v := any(v).(type) {
		case string:
			if len(v) != 4 {
				return Username{}, fmt.Errorf(
					"create username: mismatched discriminator string length, want 4, have %d",
					len(v),
				)
			}
			// Convert the string to an int
			val, err := strconv.Atoi(v)
			if err != nil {
				return Username{}, fmt.Errorf("create username: %w", err)
			}
			dI = val
		case int:
			dI = v
		default:
			panic("unreachable code")
		}

		if dI < math.MinInt16 || dI > math.MaxInt16 {
			return Username{}, fmt.Errorf(
				"create username: out of bounds integer, want [-32768,32767], have %d",
				dI,
			)
		}
		d = int16(dI)
	case int16:
		d = v
	default:
		return Username{}, fmt.Errorf("create username: invalid type `%T` (`%v`)",
			v, v,
		)
	}

	protectedHandle := slices.Contains(ReservedHandles, handle)
	vh := validHandle(handle)
	vd := validDiscriminator(d, protectedHandle)
	if !vh || !vd {
		i := make([]string, 0, 2)
		// TODO: this code still looks not good
		if !vh {
			i = append(i, "handle")
		}
		if !vd {
			i = append(i, "discriminator")
		}

		// TODO: is %v the right verb? %q?
		return Username{}, fmt.Errorf("create username: invalid %v", i)
	}

	return Username{
			handle:        handle,
			discriminator: d,
		},
		nil
}

// Generate a Username with only a handle.
// This sets the discriminator to `0000`, which is normally invalid but
// serves as a notice to the datastore to find a unique one.
func UsernameFromHandle(handle string) (Username, error) {
	if !validHandle(handle) {
		return Username{}, fmt.Errorf("invalid handle")
	}
	return Username{handle: handle, discriminator: 0}, nil
}

func (un Username) String() string {
	return fmt.Sprintf("%v#%04d",
		un.handle,
		un.discriminator)
}

func (un Username) Components() (string, int16) {
	return un.handle, un.discriminator
}

// Checks if a username is valid,
//
// This validates a Username given its constraints, this is equivalent
// to matching the PCRE2 Regular Expression:
//
// ```
// ^(?=\S)([^@#\n]{2,32})(?<=\S)#(?!0000)(\d{4})$
// ```
//
// However, due to the absence of lookarounds in Go flavour regex a
// longer process has to be performed to validate each step. This wraps
// those steps into a single method.
//
// Note that this method passing does not guarentee the exact username
// is available, merely that it is in the right form
//
// TODO: is this necessary given the tendency to validate at creation?
func (un Username) Validate(canBeZero ...bool) bool {
	return validHandle(un.handle) &&
		validDiscriminator(un.discriminator, canBeZero...)
}

func (un Username) MarshalJSON() ([]byte, error) {
	uStr := un.String()
	return json.Marshal(uStr)
}

func (un *Username) UnmarshalJSON(b []byte) error {
	var uStr string
	if err := json.Unmarshal(b, &uStr); err != nil {
		return err
	}

	if uname, err := UsernameFromString(uStr); err != nil {
		return err
	} else {
		*un = uname
		return nil
	}
}

// Wraps the handful of regular expressions necessary to validate a
// username handle into a single function.
func validHandle(handle string) bool {
	return generalHandleRe.MatchString(handle) &&
		!lHandleWhtespRe.MatchString(handle) &&
		!rHandleWhtespRe.MatchString(handle)
}

func validDiscriminator(d int16, canBeZero ...bool) bool {
	switch len(canBeZero) {
	case 0:
		break
	default:
		if canBeZero[0] {
			return d > -1 && d < 10000
		}
	}
	return d > 0 && d < 10000
}
