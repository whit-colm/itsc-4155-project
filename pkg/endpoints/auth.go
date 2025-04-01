package endpoints

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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

// TODO: do not bake secret into the f*****g source code.
var jwtSigner crypto.Signer

type CustomClaims struct {
	UserID uuid.UUID `json:"user"`
	jwt.RegisteredClaims
}

// Auth middleware for protected endpoints
func AuthenticationRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "authorization token is blank",
					Details: nil,
				},
			)
			return
		}

		// Expect header value to be in the format "Bearer <token>"
		var tokenString string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Authorization header invalid format`",
					Details: fmt.Errorf("expected: `%v`; received: `%v`",
						"Bearer ${TOKEN}",
						authHeader,
					),
				},
			)
			return
		}

		token, err := jwt.ParseWithClaims(
			tokenString,
			&CustomClaims{},
			func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodEd25519); !ok {
					return nil, jwt.ErrInvalidKey
				}
				return jwtSigner, nil
			},
		)
		// handle parse errors & invalid token
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Invalid token",
					Details: err,
				},
			)
		}

		// Verify claims
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			c.Set("userID", claims.UserID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				jsonParsableError{
					Summary: "Invalid token",
					Details: err,
				},
			)
		}
	}
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
		fmt.Println("DEBUG: CREATING NEW USER")
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
	fmt.Println("DEBUG: SEARCHING FOR USER")
	// By this point we are certain the user exists; either because
	// they've just logged in or their account has been created. We now
	// issue a JWT now.
	u, err := h.repo.GetByGithubID(c.Request.Context(), strconv.Itoa(aux.ID))
	if err != nil {
		return http.StatusServiceUnavailable,
			"failed to find",
			err
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
