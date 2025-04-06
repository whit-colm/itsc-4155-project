package endpoints

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type userHandle struct {
	repo repository.UserManager
	blob repository.BlobManager
}

var (
	uh               userHandle
	hackyUUIDtoB32RE *regexp.Regexp
)

func init() {
	// match the alphabet for `base32.StdEncoding`
	hackyUUIDtoB32RE = regexp.MustCompile(`[ABCDEFGHIJKLMNOPQRSTUVWXYZ234567]`)
}

// Handle for the signed-in user
func (h *userHandle) AccountInfo(c *gin.Context) (int, string, error) {
	id, err := wrapGinContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("get current user: %w", err)
	} else if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("get current user: %w", err)
	}
	u, err := h.repo.GetByID(c.Request.Context(), id)
	if errors.Is(err, repository.ErrorNotFound) {
		return http.StatusNotFound,
			"could not find user with ID",
			fmt.Errorf("get current user: %w", err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"unknown error looking up user",
			fmt.Errorf("get current user: %w", err)
	}
	c.JSON(http.StatusOK, u)
	return http.StatusOK, "", nil
}

// Handle for other users,
func (h *userHandle) UserInfo(c *gin.Context) (int, string, error) {
	const errorCaller string = "user info"
	paramUserID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"unable to parse parameter UUID",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	userId, err := wrapGinContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	if paramUserID == userId {
		// TODO: Make less static.
		c.Redirect(http.StatusFound, "/api/user/me")
	}
	return http.StatusFound, "", nil
}

func (h *userHandle) Delete(c *gin.Context) (int, string, error) {
	const errorCaller string = "delete user"
	// Get permissions for the current authenticated user, and the
	// value of the `id` param; these are used for administrator
	// deletion functionality.
	perms := c.GetBool("permissions")
	paramUserID, _ := uuid.Parse(c.Param("id"))

	// Get the ID of the currently authenticated user
	//
	// We do not need to handle the error because the permission check
	// middleware already requires a valid token to be passed; so if
	// we're here we know there must be one.
	tokenUser, _ := wrapGinContextUserID(c)

	// Determine the user we actually intend to delete:
	//  1. If the param ID is set we intend to delete that user; and we
	//     need to check for valid permissions
	//  2. If the param Id is nil, we intend to delete the token user.
	var userIDToDelete uuid.UUID
	if paramUserID == uuid.Nil {
		if !perms {
			return http.StatusForbidden,
				"You do not have the necessary permissions to delete other users",
				fmt.Errorf("%v: user permissions error", errorCaller)
		}
		userIDToDelete = paramUserID
	} else {
		userIDToDelete = tokenUser
	}

	// We require a query parameter called `code` so someone doesn't
	// just randomly ice their account.
	// The code is easy to calculate and is not cryptographically
	// secure. Despite it being based on TOTP it should not be used for
	// 2FA (not that we even have it given we use SSO to log in)
	// Due to time skews, we also compare it to the next and previous
	// codes.
	code := c.Query("code")
	matches := func(id uuid.UUID) []string {
		m := make([]string, 3)
		m[0] = genDeleteTOTP(id, -30*time.Second)
		m[1] = genDeleteTOTP(id, 0)
		m[2] = genDeleteTOTP(id, 30*time.Second)
		return m
	}(userIDToDelete)
	if !slices.Contains(matches, code) {
		return http.StatusForbidden,
			"Invalid deletion code, are you sure you want to do this?",
			fmt.Errorf("%v: deletion code expected `%s`, received `%s`",
				errorCaller,
				matches,
				code,
			)
	}

	err := h.repo.Delete(c.Request.Context(), userIDToDelete)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	return http.StatusOK, "", nil
}

func (h *userHandle) Update(c *gin.Context) (int, string, error) {
	const errorCaller string = "update user"
	tokenUserID, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"You must be logged in to edit your profile",
			fmt.Errorf("%v: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"Issue parsing ID from context",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	jBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusBadRequest,
			"There was an issue reading the body of your request",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	var user model.User
	if err := json.Unmarshal(jBytes, &user); err != nil {
		return http.StatusBadRequest,
			"could not parse JSON into user object",
			fmt.Errorf("%s: %w", errorCaller, err)
	}

	if user.ID == uuid.Nil {
		return http.StatusBadRequest,
			"UUID you're passing is nil, which likely means you're not sending a full object. Send a full object as the update process will automatically delete empty fields",
			fmt.Errorf("%v: patch user ID is `%v`", uuid.Nil)
	} else if user.ID != tokenUserID {
		return http.StatusForbidden,
			"You can't edit someone else's profile",
			fmt.Errorf("%v: mismatch between authenticated user ID `%v` and patch user ID `%v`",
				tokenUserID,
				user.ID,
			)
	}

	updated, err := h.repo.Update(c.Request.Context(), &user)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	c.JSON(http.StatusOK, updated)
	return http.StatusOK, "", nil
}

func (h *userHandle) UpdateAvatar(c *gin.Context) (int, string, error) {
	const errorCaller string = "update user avatar"
	tokenUser, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"You must be logged in to edit your profile",
			fmt.Errorf("%v: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"Issue parsing ID from context",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	mdata := make(map[string]string)
	if ct := c.Request.Header.Get("content-type"); strings.HasPrefix(ct, "image/") {
		mdata["content-type"] = ct
	} else {
		return http.StatusBadRequest,
			"User avatars must be an image",
			fmt.Errorf("%v: expected content-type of `image/*`, received `%v`", ct)
	}
	user, err := h.repo.GetByID(c.Request.Context(), tokenUser)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return http.StatusInternalServerError,
			"could not generate UUIDv7",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	newAvatar := &model.Blob{
		ID:       id,
		Metadata: mdata,
		Content:  c.Request.Body,
	}
	if err = h.blob.Create(c.Request.Context(), newAvatar); err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	user.Avatar = id
	user, err = h.repo.Update(c.Request.Context(), user)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, user)
	return http.StatusOK, "", nil
}

// Not an endpoint to be exposed directly!!!
func (h *userHandle) create(ctx context.Context, u model.User, a string) (int, string, error) {
	// Get the image & convert it to base64
	resp, err := http.Get(a)
	if err != nil {
		return http.StatusServiceUnavailable,
			"could not fetch profile image from URL",
			err
	}
	defer resp.Body.Close()
	mdata := make(map[string]string)
	mdata["content-type"] = resp.Header.Get("content-type")
	// This does not need to be cast to Time and back because it is
	// already a UNIX date
	mdata["last-modified"] = resp.Header.Get("last-modified")

	imgID, err := uuid.NewV7()
	if err != nil {
		return http.StatusInternalServerError,
			"could not generate UUIDv7",
			err
	}

	userID, err := uuid.NewV7()
	if err != nil {
		return http.StatusInternalServerError,
			"could not generate UUIDv7",
			err
	}

	b := model.Blob{
		ID:       imgID,
		Metadata: mdata,
		Content:  resp.Body,
	}

	u.ID = userID
	u.Avatar = imgID

	var status = http.StatusOK
	var summary string
	if err = h.blob.Create(ctx, &b); err != nil {
		// This just kinda ignores errors because not having a pfp
		// isn't the end of days.
		// So instead we just nil the field lol
		u.Avatar = uuid.Nil
		summary = "failed to commit user profile picture (this is not that bad)"
	}
	if err = h.repo.Create(ctx, &u); err != nil {
		status = http.StatusServiceUnavailable
		summary = "failed to commit new user"
	}

	return status, summary, err
}

// **NOT MEANINGFULLY SECURE**
//
// A method for generating the pin necessary for account deletion. It
// is a bodgey implementation of sorta-RFC 6238
//
// This implementation expects a 30 second rotation and 6 character len
func genDeleteTOTP(id uuid.UUID, deltaT time.Duration) string {
	// Turn a UUID into all caps and strip all characters that are not
	// Base32.
	secret := strings.ToUpper(id.String())
	secret = strings.Join(
		hackyUUIDtoB32RE.FindAllString(secret, -1),
		"",
	)

	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		return err.Error()
	}

	// Get the current time + offset as unix timestamp (int64)
	now := time.Now().Add(deltaT).Unix()
	count := uint64(now / 30)

	var countByes [8]byte
	for i := 7; i >= 0; i-- {
		countByes[i] = byte(count & 0xff)
		count >>= 8
	}

	// Now on to doing HMAC things
	h := hmac.New(sha1.New, key)
	h.Write(countByes[:])
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	code := (uint32(hash[offset])&0x7f)<<24 |
		(uint32(hash[offset+1])&0xff)<<16 |
		(uint32(hash[offset+2])&0xff)<<8 |
		(uint32(hash[offset+3]) & 0xff)

	// Hacky anonymous function for raising an integer to an integer
	// power
	powi := func(n int) int {
		res := 1
		for range n {
			res *= 10
		}
		return res
	}

	otp := code % uint32(powi(6))

	return fmt.Sprintf("%06d", otp)
}
