package endpoints

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

/* Much of this code and documentation was written with reference to Eli
 * Bendersky's blog post.
 *
 * https://eli.thegreenplace.net/2023/sign-in-with-github-in-go/
 */

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")

		if user == nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

type userHandle struct {
	repo repository.UserManager
	blob repository.BlobManager
}

// Initial login endpoint for the user
func (h *userHandle) Login(c *gin.Context) {
	// Generate state
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
func (h *userHandle) GithubCallback(c *gin.Context) {
	if state, err := c.Cookie("state"); err != nil {
		fmt.Printf("%v:\t%v\n", state, err)
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "state not found",
				Details: err})
		return
	} else if c.Query("state") != state {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "state did not match",
				Details: err})
		return
	}

	tok, err := conf.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "token exchange failure",
				Details: err})
		return
	}

	client := conf.Client(c.Request.Context(), tok)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "could not generate client from token",
				Details: err})
		return
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
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "could not read body",
				Details: err})
		return
	} else if err = json.Unmarshal(rb, &aux); err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "could not unmarshal user info",
				Details: err})
		return
	}

	// Now we have to either create a user if this is the first log-in
	// otherwise we can just issue our token and move on
	if exists, err := h.repo.ExistsByGithubID(c.Request.Context(), string(aux.ID)); err != nil {
		summary := fmt.Sprintf("could not retrieve user with ID `%s`", string(aux.ID))
		c.JSON(http.StatusServiceUnavailable,
			jsonParsableError{Summary: summary,
				Details: err})
		return
	} else if !exists {
		// Get the image & convert it to base64
		imgB64, code, jErr := func(url string) (string, int, *jsonParsableError) {
			resp, err := http.Get(url)
			defer resp.Body.Close()
			if err != nil {
				return "",
					http.StatusServiceUnavailable,
					&jsonParsableError{
						Summary: "could not fetch profile image from URL",
						Details: err,
					}
			}

			var imgB64 string
			if rb, err := io.ReadAll(resp.Body); err != nil {
				return "",
					http.StatusInternalServerError,
					&jsonParsableError{
						Summary: "could not read body for profile image",
						Details: err,
					}
			} else {
				imgB64 = base64.URLEncoding.EncodeToString(rb)
			}

			return imgB64, 200, nil
		}(aux.AvatarURL)
		if jErr != nil {
			c.JSON(code, *jErr)
		}

		imgID, err := uuid.NewV7()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				jsonParsableError{Summary: "could not generate UUIDv7",
					Details: err})
			return
		}

		userID, err := uuid.NewV7()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				jsonParsableError{Summary: "could not generate UUIDv7",
					Details: err})
			return
		}

		b := model.Blob{
			ID:      imgID,
			Content: imgB64,
		}

		u := model.User{
			ID:          userID,
			GitHubID:    string(aux.ID),
			DisplayName: aux.Name,
			UserHandle:  aux.Login,
			Email:       aux.Email,
			Avatar:      imgID,
		}
		h.blob.Create(c.Request.Context(), &b)
		h.repo.Create(c.Request.Context(), &u)
	}

	c.JSONP(http.StatusOK, aux)
}

func (h *userHandle) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, "{}")
}

func (h *userHandle) AccountInfo(c *gin.Context) {
	c.JSON(http.StatusOK, "{}")
}
