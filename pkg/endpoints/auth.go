package endpoints

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

/* Much of this code and documentation was written with reference to Eli
 * Bendersky's blog post.
 *
 * https://eli.thegreenplace.net/2023/sign-in-with-github-in-go/
 */

type jwtCustomSigner struct {
	priv crypto.Signer
	pub  crypto.PublicKey
}

var (
	jwtSigner            jwtCustomSigner
	errUserIDKeyNotFound = errors.New("could not find key `UserID` in context")
)

// test if custom signer implements inheritance
var _ crypto.Signer = (*jwtCustomSigner)(nil)

func (j jwtCustomSigner) Public() crypto.PublicKey {
	return j.pub
}

func (j jwtCustomSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return j.priv.Sign(rand, digest, opts)
}

// Auth middleware to validate token & set up user-specific data in gontext
//
// This does *NOT* validate if the user can access a resource, it
// merely decrypts user data from the authorization token.
//
// There are three defined behaviors:
//
//  1. If no authorization token is passed it continues without
//     modifying the gin context
//  2. If an authorization token is passed and can be validated, it
//     stores the user's UUID in the gin context's `userID` key
//  3. If an authorization token is passed but cannot be validated, it
//     aborts with a JSON status
func AuthorizationJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// Yes "Bearer null" is not... something I should check for
		// explicitly, but I am already asking a lot of the frontend
		// girls so I'll just make this a nonissue for them
		if authHeader == "" || authHeader == "Bearer null" {
			c.Next()
			return
		}

		// Expect header value to be in the format "Bearer <token>"
		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				jsonParsableError{
					Summary: "authorization header invalid format`",
					Details: fmt.Errorf("authorization header expected `%v`, received `%v`",
						"Bearer ${TOKEN}",
						authHeader,
					),
				},
			)
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&jwt.RegisteredClaims{},
			func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
					return nil, jwt.ErrInvalidKey
				}
				return jwtSigner.Public(), nil
			},
		)
		// handle parse errors & invalid token
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Failed to parse token claims.",
					Details: fmt.Errorf("get authorization JWT: %w", err),
				},
			)
		}

		// Verify claims
		if subj, err := token.Claims.GetSubject(); err == nil && token.Valid {
			c.Set("userID", subj)
			c.Next()
		} else if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Token is invalid.",
					Details: fmt.Errorf("get authorization JWT: %w", err),
				},
			)
		} else {
			c.AbortWithStatusJSON(http.StatusBadRequest,
				jsonParsableError{
					Summary: "Get claim subject returned error.",
					Details: fmt.Errorf("get authorization JWT: %w", err),
				},
			)
		}
	}
}

// UserPermissions is a function which informs the context of the
// requesting user's permissions. This requires that some authorization
// has been done before hand and a valid user ID stored in the gontext
// map as `"userID"`
func UserPermissions() gin.HandlerFunc {
	const errorCaller string = "check user permissions"
	return func(c *gin.Context) {
		id, err := wrapGinContextUserID(c)
		if errors.Is(err, errUserIDKeyNotFound) {
			c.Set("permissions", false)
			c.Next()
			return
		} else if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Issue parsing ID from context, are you sure you're authenticated?",
					Details: fmt.Errorf("%v: %w", errorCaller, err),
				})
			return
		}
		admin, err := ah.repo.Permissions(c.Request.Context(), id)
		if err != nil {
			h, s, d := wrapDatastoreError(errorCaller, err)
			c.AbortWithStatusJSON(h, jsonParsableError{s, d})
			return
		}
		c.Set("permissions", admin)
		c.Next()
	}
}

// Wrapper to get usable UUID type from gin context key-value store
func wrapGinContextUserID(c *gin.Context) (uuid.UUID, error) {
	idAny, ok := c.Get("userID")
	if !ok {
		return uuid.Nil, errUserIDKeyNotFound
	}
	idStr, valid := idAny.(string)
	if !valid {
		return uuid.Nil, fmt.Errorf("assert %t `%#v` (any) -> `string`",
			valid, idAny)
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("convert %s (string) -> `UUID`: %w",
			idStr, err)
	}
	return id, nil
}

/** Authentication endpoints **/

type authHandle struct {
	repo repository.UserManager
	auth repository.AuthManager
}

var ah authHandle

// Initial login endpoint for the user
func (h *authHandle) Login(c *gin.Context) {
	// Generate state, used to verify the user sent back by GitHub is
	// the same one we sent it. It's just a random string.
	state, err := func(n int) (string, error) {
		bytes := make([]byte, n)
		if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
			return "", err
		}
		return base64.URLEncoding.EncodeToString(bytes), nil
	}(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "could not generate state",
				Details: err})
		return
	}

	c.SetCookie(
		"state",
		state,
		int(time.Hour.Seconds()),
		"/",
		"localhost",
		c.Request.TLS != nil,
		true,
	)

	c.Redirect(http.StatusFound, conf.AuthCodeURL(state))
}

// The redirection path GitHub yaps at
//
// By now, GitHub has authenticated the user, we can exchange the
func (h *authHandle) GithubCallback(c *gin.Context) (int, string, error) {
	if state, err := c.Cookie("state"); err != nil {
		return http.StatusBadRequest,
			"state not found",
			err
	} else if c.Query("state") != state {
		return http.StatusBadRequest,
			"state did not match",
			err
	}

	tok, err := conf.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		return http.StatusBadRequest,
			"token exchange failure",
			err
	}

	client := conf.Client(c.Request.Context(), tok)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return http.StatusInternalServerError,
			"could not generate client from token",
			err
	}
	defer resp.Body.Close()

	aux := struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	}{}

	// respbody is our JSON object defining the user
	if rb, err := io.ReadAll(resp.Body); err != nil {
		return http.StatusInternalServerError,
			"could not read body",
			err
	} else if err = json.Unmarshal(rb, &aux); err != nil {
		return http.StatusInternalServerError,
			"could not unmarshal user info",
			err
	}

	// Now we have to either create a user if this is the first log-in
	// otherwise we can just issue our token and move on
	if exists, err := h.repo.ExistsByGithubID(c.Request.Context(), strconv.Itoa(aux.ID)); err != nil {
		return http.StatusServiceUnavailable,
			fmt.Sprintf("could not retrieve user with ID `%s`", strconv.Itoa(aux.ID)),
			err
	} else if !exists {
		handle, err := model.UsernameFromHandle(aux.Login)
		if err != nil {
			return http.StatusInternalServerError,
				"invalid handle, could not coerce username",
				err
		}
		// Generate incomplete User to create
		// (everything else done in userhandle create method)
		u := model.User{
			GithubID:    strconv.Itoa(aux.ID),
			DisplayName: aux.Name,
			Username:    handle,
			Email:       aux.Email,
			Admin:       false,
		}

		if status, summary, serr := uh.create(c.Request.Context(), u, aux.AvatarURL); status != http.StatusOK {
			return status, summary, serr
		} else if serr != nil {
			c.Error(serr)
		}
	}
	// By this point we are certain the user exists; either because
	// they've just logged in or their account has been created. We now
	// issue a JWT now.
	u, err := h.repo.GetByGithubID(c.Request.Context(), strconv.Itoa(aux.ID))
	if errors.Is(err, repository.ErrNotFound) {
		return http.StatusInternalServerError,
			"Did not find user with the given GitHub ID",
			fmt.Errorf("login github callback: %w", err)
	} else if errors.Is(err, repository.ErrBadConnection) {
		return http.StatusServiceUnavailable,
			"Issue querying the database",
			fmt.Errorf("login github callback: %w", err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"Error searching for user with the given GitHub ID",
			fmt.Errorf("login github callback: %w", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    "JAWS_test_app",
		Subject:   u.ID.String(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	if tokenString, err := token.SignedString(jwtSigner); err != nil {
		return http.StatusInternalServerError,
			"could not generate token",
			err
	} else {
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
		return 0, "", nil
	}
}

func (h *authHandle) Logout(c *gin.Context) (code int, sum string, err error) {
	c.JSON(http.StatusOK, "{}")
	return
}
