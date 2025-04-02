package endpoints

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"errors"
	"fmt"
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
	id, err := ginContextUserID(c)
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
	paramId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"unable to parse parameter UUID",
			fmt.Errorf("user info: %w", err)
	}
	userId, err := ginContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("get current user: %w", err)
	}
	if paramId == userId {
		// TODO: Make less static.
		c.Redirect(http.StatusFound, "/api/user/me")
	}
	panic("unimplemented!")
}

func (h *userHandle) Delete(c *gin.Context) {
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
		m[0] = genDeleteTOTP(uuid.Nil, -30*time.Second)
		m[1] = genDeleteTOTP(uuid.Nil, 0)
		m[2] = genDeleteTOTP(uuid.Nil, 30*time.Second)
		return m
	}(uuid.Nil)
	if !slices.Contains(matches, code) {
		c.JSON(http.StatusForbidden,
			jsonParsableError{
				Summary: "Invalid deletion code, are you sure you want to do this?",
				Details: fmt.Errorf("expected: `%s`; received: `%s`",
					matches,
					code),
			},
		)
		return
	}

	err := h.repo.Delete(c.Request.Context(), uuid.Nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{
				Summary: "deleting user failed",
				Details: err,
			},
		)
	}
	// Try to delete but we don't really care tbh
	h.blob.Delete(c.Request.Context(), uuid.Nil)
}

func (h *userHandle) Update(ctx *gin.Context) {
	panic("unimplemented")
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
		ID:      imgID,
		Content: resp.Body,
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
