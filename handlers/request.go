package handlers

import (
	"net/http"
	"strings"

	"github.com/caiyeon/goldfish/request"
	"github.com/caiyeon/goldfish/slack"
	"github.com/caiyeon/goldfish/vault"
	"github.com/labstack/echo"
)

// Finds a request and returns the details, if the user has vault permissions to read the resources
func GetRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// fetch request from cubbyhole
		req, err := request.Get(auth, c.FormValue("hash"))
		if err != nil {
			// if error contains 403 from vault, forward it to the user
			if strings.Contains(err.Error(), "Code: 403. Errors:\n\n* permission denied") {
				return c.JSON(http.StatusForbidden, H{
					"error": err.Error(),
				})
			} else {
				return c.JSON(http.StatusBadRequest, H{
					"error": err.Error(),
				})
			}
		}

		// return request details
		return c.JSON(http.StatusOK, H{
			"result": req,
			"error":  "",
		})
	}
}

// Adds a request to cubbyhole, that can be rejected/approved later
// Requires requester to have read access to the policy
func AddRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// bind body to an arbitrary map. Requests package should take care of interpretation
		params := make(map[string]interface{})
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Body must be in JSON format",
			})
		}

		// type field is required
		if _, exists := params["Type"]; !exists {
			if _, exists = params["type"]; !exists {
				return c.JSON(http.StatusBadRequest, H{
					"error": "'Type' field is required in request body",
				})
			}
		}

		// add the request
		hash, err := request.Add(auth, params)
		if err != nil {
			// if error contains 403 from vault, forward it to the user
			if strings.Contains(err.Error(), "Code: 403. Errors:\n\n* permission denied") {
				return c.JSON(http.StatusForbidden, H{
					"error": err.Error(),
				})
			} else {
				return c.JSON(http.StatusBadRequest, H{
					"error": err.Error(),
				})
			}
		}

		// if config has a slack webhook, send the hash (aka change ID) to the channel
		conf := vault.GetConfig()
		if conf.SlackWebhook != "" {
			// send a message using webhook
			err = slack.PostMessageWebhook(
				conf.SlackChannel,
				"A new policy change request has been submitted",
				"Request ID: \n*"+hash+"*",
				conf.SlackWebhook,
			)
			// change request is fine, just let the frontend know it wasn't slack'd
			if err != nil {
				return c.JSON(http.StatusOK, H{
					"result": hash,
					"error":  "Could not send to slack webhook",
				})
			}
		}

		// if all is good, return hash
		return c.JSON(http.StatusOK, H{
			"result": hash,
			"error":  "",
		})
	}
}

func ApproveRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		// ensure unseal key is provided in json body
		params := make(map[string]interface{})
		if err := c.Bind(&params); err != nil {
			return c.JSON(http.StatusBadRequest, H{
				"error": "Body must be in JSON format",
			})
		}
		unseal, exists := params["unseal"]
		if !exists || unseal == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "'unseal' parameter is required",
			})
		}
		hash, exists := params["hash"]
		if !exists {
			hash = c.FormValue("hash")
		}
		if hash == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "'hash' parameter is required",
			})
		}

		// approve the request by hash
		req, err := request.Approve(auth, hash.(string), unseal.(string))
		if err != nil {
			// if error contains 403 from vault, forward it to the user
			if strings.Contains(err.Error(), "Code: 403. Errors:\n\n* permission denied") {
				return c.JSON(http.StatusForbidden, H{
					"error": err.Error(),
				})
			} else {
				return c.JSON(http.StatusBadRequest, H{
					"error": err.Error(),
				})
			}
		}

		return c.JSON(http.StatusOK, H{
			"result": req,
		})
	}
}

func RejectRequest() echo.HandlerFunc {
	return func(c echo.Context) error {
		// fetch auth from header
		auth := getSession(c)
		if auth == nil {
			return nil
		}
		defer auth.Clear()

		hash := c.FormValue("hash")
		if hash == "" {
			return c.JSON(http.StatusBadRequest, H{
				"error": "'hash' parameter is required",
			})
		}

		if err := request.Reject(auth, hash); err != nil {
			// if error contains 403 from vault, forward it to the user
			if strings.Contains(err.Error(), "Code: 403. Errors:\n\n* permission denied") {
				return c.JSON(http.StatusForbidden, H{
					"error": err.Error(),
				})
			} else {
				return c.JSON(http.StatusBadRequest, H{
					"error": err.Error(),
				})
			}
		}

		return c.JSON(http.StatusOK, H{
			"result": "success",
		})
	}
}
